package proxy

import (
	"fmt"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/miekg/dns"
)

type NameProxy struct {
	listen string
	cfg    *config.NameProxy
	server *dns.Server
	out    *libol.SubLogger
	lock   sync.RWMutex
	names  map[string]string
	addrs  map[string]string
}

func NewNameProxy(cfg *config.NameProxy) *NameProxy {
	return &NameProxy{
		listen: cfg.Listen,
		cfg:    cfg,
		out:    libol.NewSubLogger(cfg.Listen),
		names:  make(map[string]string),
		addrs:  make(map[string]string),
	}
}

func (n *NameProxy) Initialize() {
}

func (n *NameProxy) Forward(name, addr, nexthop string) {
	opts := []string{"metric", fmt.Sprintf("%d", n.cfg.Metric)}
	if out, err := libol.IpRouteAdd("", addr, nexthop, opts...); err != nil {
		n.out.Warn("Access.Forward: %s %s: %s", addr, err, out)
		return
	}
	n.out.Info("NameProxy.Forward: %s <- %s via %s ", nexthop, name, addr)
}

func (n *NameProxy) UpdateDNS(name, addr string) bool {
	n.lock.Lock()
	defer n.lock.Unlock()

	updated := false
	if _, ok := n.names[name]; !ok {
		n.names[name] = addr
		updated = true
	}

	if _, ok := n.addrs[addr]; !ok {
		n.addrs[addr] = name
		updated = true
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
	client := &dns.Client{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Net:          "udp",
	}
	nameto := n.cfg.Nameto

	libol.Go(func() {
		via := n.FindBackend(r)

		if via != nil {
			nameto = via.Nameto
		}

		config.SetListen(&nameto, 53)
		if nameto == "0.0.0.0:53" || nameto == n.listen {
			n.out.Error("NameProxy.handleDNS nil(%s)", nameto)
			return
		}

		n.out.Info("NameProxy.handleDNS %s <- %v via %s", nameto, r.Question, conn.RemoteAddr())
		resp, _, err := client.Exchange(r, nameto)
		if err != nil {
			n.out.Error("NameProxy.handleDNS %s: %v", r, err)
			return
		}

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
	dns.HandleFunc(".", n.handleDNS)
	n.server = &dns.Server{Addr: n.listen, Net: "udp"}
	n.out.Info("NameProxy.StartDNS on %s", n.listen)

	if err := n.server.ListenAndServe(); err != nil {
		n.out.Error("NameProxy.StartDNS server: %v", err)
	}
}

func (n *NameProxy) Stop() {
	if n.server != nil {
		n.server.Shutdown()
		n.server = nil
	}
	n.out.Info("NameProxy.Stop")
}

func (n *NameProxy) Save() {
}
