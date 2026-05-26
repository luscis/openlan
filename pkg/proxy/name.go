package proxy

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/access"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
	"github.com/miekg/dns"
)

const CacheTimeout = 5
const DefaultCacheTimeout = 2 * time.Minute
const CacheLogInterval = 5 * time.Second

type AddrCache struct {
	Address string
	Time    int64
}

func (a *AddrCache) Expire() bool {
	dt := time.Now().Unix() - a.Time
	if dt > CacheTimeout {
		return true
	}
	return false
}

func (a *AddrCache) Update() {
	a.Time = time.Now().Unix()
}

type NameCache struct {
	Name string
	Time int64
}

func (a *NameCache) Expire() bool {
	dt := time.Now().Unix() - a.Time
	if dt > CacheTimeout {
		return true
	}
	return false
}

func (a *NameCache) Update() {
	a.Time = time.Now().Unix()
}

type NameProxy struct {
	listen string
	cfg    *config.NameProxy
	server *dns.Server
	out    *libol.SubLogger
	lock   sync.RWMutex
	names  map[string]*AddrCache
	addrs  map[string]*NameCache
	access []*access.Access
	rlLock sync.Mutex
	rlTick int64
	rlStat map[string]int
	dnsMsg map[string]*DNSMsgCache
	chLock sync.Mutex
	chStat map[string]*cacheHitStat
}

type DNSMsgCache struct {
	msg  *dns.Msg
	time time.Time
}

type cacheHitStat struct {
	tick time.Time
	cnt  int
}

func NewNameProxy(cfg *config.NameProxy) *NameProxy {
	n := &NameProxy{
		listen: cfg.Listen,
		cfg:    cfg,
		out:    libol.NewSubLogger(cfg.Listen),
		names:  make(map[string]*AddrCache),
		addrs:  make(map[string]*NameCache),
		rlStat: make(map[string]int),
		dnsMsg: make(map[string]*DNSMsgCache),
		chStat: make(map[string]*cacheHitStat),
	}
	n.Initialize()
	return n
}

func (n *NameProxy) findKey(r *dns.Msg) string {
	if len(r.Question) != 1 {
		return ""
	}
	q := r.Question[0]
	return fmt.Sprintf("%s|%d|%d", q.Name, q.Qtype, q.Qclass)
}

func (n *NameProxy) findCache(key string) *dns.Msg {
	if !n.dnsCacheEnabled() {
		return nil
	}
	if key == "" {
		return nil
	}
	n.lock.RLock()
	entry, ok := n.dnsMsg[key]
	n.lock.RUnlock()
	if !ok {
		return nil
	}

	ttl := n.dnsCacheTimeout()
	if time.Since(entry.time) > ttl {
		n.lock.Lock()
		if latest, ok := n.dnsMsg[key]; ok && time.Since(latest.time) > ttl {
			delete(n.dnsMsg, key)
		}
		n.lock.Unlock()
		return nil
	}

	return entry.msg.Copy()
}

func (n *NameProxy) dnsCacheTimeout() time.Duration {
	if n.cfg == nil || n.cfg.CacheTTL < 0 {
		return DefaultCacheTimeout
	}
	return time.Duration(n.cfg.CacheTTL) * time.Second
}

func (n *NameProxy) dnsCacheEnabled() bool {
	return n.cfg != nil && n.cfg.CacheTTL != 0
}

func (n *NameProxy) updateCache(key string, msg *dns.Msg) {
	if !n.dnsCacheEnabled() {
		return
	}
	if key == "" || msg == nil {
		return
	}
	n.lock.Lock()
	n.dnsMsg[key] = &DNSMsgCache{
		msg:  msg.Copy(),
		time: time.Now(),
	}
	n.lock.Unlock()
}

func (n *NameProxy) logCache(conn dns.ResponseWriter, r *dns.Msg) {
	addr := libol.GetIPAddr(conn.RemoteAddr().String())
	now := time.Now()
	needLog := false
	count := 0

	n.chLock.Lock()
	stat, ok := n.chStat[addr]
	if !ok {
		stat = &cacheHitStat{tick: now}
		n.chStat[addr] = stat
	}
	if now.Sub(stat.tick) >= CacheLogInterval {
		count = stat.cnt
		stat.cnt = 0
		stat.tick = now
		needLog = true
	}
	stat.cnt++
	n.chLock.Unlock()

	if needLog && count > 0 {
		n.out.Info("NameProxy.handleDNS cache hit client=%s %d/%ds latest=%v", addr, count, int(CacheLogInterval/time.Second), r.Question)
	}
}

func (n *NameProxy) allowDNS(conn dns.ResponseWriter) bool {
	if n.cfg.Rate <= 0 {
		return true
	}

	addr := libol.GetIPAddr(conn.RemoteAddr().String())
	now := time.Now().Unix()

	n.rlLock.Lock()
	defer n.rlLock.Unlock()

	if now != n.rlTick {
		n.rlTick = now
		n.rlStat = make(map[string]int)
	}

	n.rlStat[addr]++
	return n.rlStat[addr] <= n.cfg.Rate
}

func (n *NameProxy) Initialize() {
	for _, cfg := range n.cfg.Access {
		acc := access.NewAccess(cfg)
		acc.Initialize()
		n.access = append(n.access, acc)
	}
}

func (n *NameProxy) Forward(name, addr, nexthop string) {
	opts := []string{}
	if runtime.GOOS == "linux" {
		opts = []string{"metric", fmt.Sprintf("%d", n.cfg.Metric)}
	}

	if strings.HasPrefix(nexthop, "local") {
		n.out.Info("NameProxy.Forward: %s <- %s via %s ", nexthop, name, addr)
		return
	}
	if out, err := network.RouteAdd("", addr, nexthop, opts...); err != nil {
		n.out.Warn("Access.Forward: %s %s: %s", addr, err, out)
		return
	}
	n.out.Info("NameProxy.Forward: %s <- %s via %s ", nexthop, name, addr)
}

func (n *NameProxy) UpdateDNS(name, addr string) bool {
	n.lock.Lock()
	defer n.lock.Unlock()

	updated := false
	if o, ok := n.names[name]; !ok {
		n.names[name] = &AddrCache{
			Address: addr,
			Time:    time.Now().Unix(),
		}
		updated = true
	} else if o.Expire() {
		updated = true
		o.Update()
	}
	if o, ok := n.addrs[addr]; !ok {
		n.addrs[addr] = &NameCache{
			Name: name,
			Time: time.Now().Unix(),
		}
		updated = true
	} else if o.Expire() {
		updated = true
		o.Update()
	}
	return updated
}

func (n *NameProxy) FindBackend(r *dns.Msg) *config.ForwardTo {
	if len(r.Question) == 0 {
		return nil
	}

	name := r.Question[0].Name
	n.out.Debug("NameProxy.FindBackend %s", name)

	n.lock.RLock()
	defer n.lock.RUnlock()

	via := n.cfg.Backends.FindBackend(name)
	if via != nil {
		n.out.Debug("NameProxy.FindBackend %s via %s", name, via.Server)
	}
	return via
}

func (n *NameProxy) handleDNS(conn dns.ResponseWriter, r *dns.Msg) {
	if !n.allowDNS(conn) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Rcode = dns.RcodeRefused
		if err := conn.WriteMsg(msg); err != nil {
			n.out.Warn("NameProxy.handleDNS ratelimit write: %s", err)
		}
		n.out.Warn("NameProxy.handleDNS drop %s over %d requests/s", conn.RemoteAddr(), n.cfg.Rate)
		return
	}

	client := &dns.Client{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Net:          "udp",
	}
	nameto := n.cfg.Nameto

	libol.Go(func() {
		via := n.FindBackend(r)
		if via != nil && via.Nameto != "" { // Override nameto if backend is found.
			nameto = via.Nameto
		}

		config.SetListen(&nameto, 53)
		if nameto == "0.0.0.0:53" || nameto == n.listen {
			n.out.Error("NameProxy.handleDNS nil(%s)", nameto)
			return
		}

		key := n.findKey(r)
		if cached := n.findCache(key); cached != nil {
			cached.Id = r.Id
			cached.Question = r.Question
			if err := conn.WriteMsg(cached); err != nil {
				n.out.Error("NameProxy.handleDNS write cache: %s", err)
			}
			n.logCache(conn, r)
			return
		}

		n.out.Info("NameProxy.handleDNS %s <- %v via %s", nameto, r.Question, conn.RemoteAddr())
		resp, _, err := client.Exchange(r, nameto)
		if err != nil {
			n.out.Error("NameProxy.handleDNS %s: %v", r, err)
			return
		}
		n.updateCache(key, resp)

		if via != nil && via.Server != "" {
			for _, rr := range resp.Answer {
				if aa, ok := rr.(*dns.A); ok {
					name := aa.Hdr.Name
					addr := aa.A.String()
					if n.UpdateDNS(name, addr) {
						n.Forward(name, addr, via.Server)
					}
				}
			}
		}

		if err := conn.WriteMsg(resp); err != nil {
			n.out.Error("NameProxy.handleDNS %s", err)
		}
	})
}

func (n *NameProxy) Start() {
	// Forward Name Server to backend nexthop server.
	n.cfg.Backends.List(func(ft *config.ForwardTo) {
		if addr := libol.GetIPAddr(ft.Nameto); addr != "" {
			n.out.Info("NameProxy.Backend %s via %s", addr, ft.Server)
			n.Forward("", addr, ft.Server)
		}
	})

	dns.HandleFunc(".", n.handleDNS)
	n.server = &dns.Server{Addr: n.listen, Net: "udp"}

	n.out.Info("NameProxy.StartDNS on %s", n.listen)

	for _, acc := range n.access {
		libol.Go(acc.Start)
	}

	if err := n.server.ListenAndServe(); err != nil {
		n.out.Error("NameProxy.StartDNS server: %v", err)
	}
}

func (n *NameProxy) Stop() {
	for _, acc := range n.access {
		acc.Stop()
	}
	n.access = nil
	if n.server != nil {
		n.server.Shutdown()
		n.server = nil
	}
	n.out.Info("NameProxy.Stop")
}

func (n *NameProxy) Save() {
}
