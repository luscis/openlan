package access

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type jobTimer struct {
	Time int64
	Call func() error
}

type KeepAlive struct {
	Interval int64
	LastTime int64
}

func (k *KeepAlive) Should() bool {
	return time.Now().Unix()-k.LastTime >= k.Interval
}

func (k *KeepAlive) Update() {
	k.LastTime = time.Now().Unix()
}

var (
	EvSocConed   = "conned"
	EvSocRecon   = "reconn"
	EvSocClosed  = "closed"
	EvSocSuccess = "success"
	EvSocSignIn  = "signIn"
	EvSocLogin   = "login"
	EvTapIpAddr  = "ipAddr"
	EvTapReadErr = "readErr"
	EvTapReset   = "reset"
	EvTapOpenErr = "openErr"
)

type WorkerEvent struct {
	Type   string
	Reason string
	Time   int64
	Data   interface{}
}

func (e *WorkerEvent) String() string {
	return e.Type + " " + e.Reason
}

func NewEvent(newType, reason string) *WorkerEvent {
	return &WorkerEvent{
		Type:   newType,
		Time:   time.Now().Unix(),
		Reason: reason,
	}
}

type WorkerListener struct {
	AddAddr   func(ipStr, gateway string) error
	DelAddr   func(ipStr string) error
	OnTap     func(w *TapWorker) error
	AddRoutes func(routes []*models.Route) error
	DelRoutes func(routes []*models.Route) error
	Forward   func(name, prefix, nexthop string)
}

type PrefixRule struct {
	Type        int
	Destination net.IPNet
	NextHop     net.IP
}

func GetSocketClient(p *config.Access, remote string) libol.SocketClient {
	crypt := p.Crypt
	block := libol.NewBlockCrypt(crypt.Algo, crypt.Secret)

	if remote == "" {
		remote = p.Connection
	}
	switch p.Protocol {
	case "kcp":
		c := libol.NewKcpConfig()
		c.Block = block
		c.RdQus = p.Queue.SockRd
		c.WrQus = p.Queue.SockWr
		return libol.NewKcpClient(remote, c)
	case "tcp":
		c := &libol.TcpConfig{
			Block: block,
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		return libol.NewTcpClient(remote, c)
	case "udp":
		c := &libol.UdpConfig{
			Block:   block,
			Timeout: time.Duration(p.Timeout) * time.Second,
			RdQus:   p.Queue.SockRd,
			WrQus:   p.Queue.SockWr,
		}
		return libol.NewUdpClient(remote, c)
	case "ws":
		c := &libol.WebConfig{
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		return libol.NewWebClient(remote, c)
	case "wss":
		c := &libol.WebConfig{
			Block: block,
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		if p.Cert != nil {
			c.Cert = &libol.CertConfig{
				Insecure: p.Cert.Insecure,
				RootCa:   p.Cert.CaFile,
			}
		}
		return libol.NewWebClient(remote, c)
	default:
		c := &libol.TcpConfig{
			Block: block,
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		if p.Cert != nil {
			c.Tls = &tls.Config{
				InsecureSkipVerify: p.Cert.Insecure,
				RootCAs:            p.Cert.GetCertPool(),
			}
		}
		return libol.NewTcpClient(remote, c)
	}
}

func GetTapCfg(c *config.Access) network.TapConfig {
	cfg := network.TapConfig{
		Provider: c.Interface.Provider,
		Name:     c.Interface.Name,
		Network:  c.Interface.Address,
		KernBuf:  c.Queue.VirSnd,
		VirBuf:   c.Queue.VirWrt,
	}
	if c.Interface.Provider == "tun" {
		cfg.Type = network.TUN
	} else {
		cfg.Type = network.TAP
	}
	return cfg
}

type Worker struct {
	// private
	ifAddr    string
	listener  WorkerListener
	conWorker *SocketWorker
	sosWorker []*SocketWorker
	tapWorker *TapWorker
	cfg       *config.Access
	uuid      string
	network   *models.Network
	routes    map[string]PrefixRule
	out       *libol.SubLogger
	done      chan bool
	ticker    *time.Ticker
	nameCache map[string]string
	addrCache map[string]string
	lock      sync.RWMutex
}

func NewWorker(cfg *config.Access) *Worker {
	return &Worker{
		ifAddr:    cfg.Interface.Address,
		cfg:       cfg,
		routes:    make(map[string]PrefixRule),
		out:       libol.NewSubLogger(cfg.Id()),
		done:      make(chan bool),
		ticker:    time.NewTicker(2 * time.Second),
		nameCache: make(map[string]string),
		addrCache: make(map[string]string),
		sosWorker: make([]*SocketWorker, 2),
	}
}

func (w *Worker) Initialize() {
	if w.cfg == nil {
		return
	}
	pid := os.Getpid()
	if fp, err := libol.OpenWrite(w.cfg.PidFile); err == nil {
		_, _ = fp.WriteString(fmt.Sprintf("%d", pid))
	}
	w.out.Info("Worker.Initialize")
	client := GetSocketClient(w.cfg, "")
	w.conWorker = NewSocketWorker(client, w.cfg)
	if w.cfg.Fallback != "" {
		back := GetSocketClient(w.cfg, w.cfg.Fallback)
		w.sosWorker[1] = NewSocketWorker(back, w.cfg)
	}
	w.sosWorker[0] = w.conWorker

	tapCfg := GetTapCfg(w.cfg)
	// register listener
	w.tapWorker = NewTapWorker(tapCfg, w.cfg)

	for _, conn := range w.sosWorker {
		conn.SetUUID(w.UUID())
		conn.listener = SocketWorkerListener{
			OnClose:   w.OnClose,
			OnSuccess: w.OnSuccess,
			OnIpAddr:  w.OnIpAddr,
			ReadAt:    w.tapWorker.Write,
		}
		conn.Initialize()
	}

	w.tapWorker.listener = TapWorkerListener{
		OnOpen: func(t *TapWorker) error {
			if w.listener.OnTap != nil {
				if err := w.listener.OnTap(t); err != nil {
					return err
				}
			}
			if w.network != nil {
				n := w.network
				// remove older firstly
				w.FreeIpAddr()
				_ = w.OnIpAddr(w.conWorker, n)
			}
			return nil
		},
		ReadAt: func(frame *libol.FrameMessage) error {
			for _, conn := range w.sosWorker {
				if !conn.client.Have(libol.ClAuth) {
					continue
				}
				return conn.Write(frame)
			}
			return nil
		},
		FindNext: w.FindNext,
	}
	w.tapWorker.Initialize()
}

func (w *Worker) SaveStatus() {
	w.lock.RLock()
	defer w.lock.RUnlock()

	file := w.cfg.StatusFile
	device := w.tapWorker.device
	client := w.conWorker.client
	if file == "" || device == nil || client == nil {
		return
	}

	sts := client.Statistics()
	access := &schema.Access{
		RxBytes:   uint64(sts[libol.CsRecvOkay]),
		TxBytes:   uint64(sts[libol.CsSendOkay]),
		ErrPkt:    uint64(sts[libol.CsSendError]),
		Uptime:    client.UpTime(),
		State:     client.Status().String(),
		Device:    device.Name(),
		Network:   w.cfg.Network,
		Protocol:  w.cfg.Protocol,
		User:      strings.SplitN(w.cfg.Username, "@", 2)[0],
		Remote:    w.cfg.Connection,
		AliveTime: client.AliveTime(),
		UUID:      w.uuid,
		Alias:     w.cfg.Alias,
		System:    runtime.GOOS,
		Names:     w.addrCache,
	}
	if w.network != nil {
		access.Address = w.network.Address
	}

	libol.MarshalSave(access, file, true)
}

func (w *Worker) Start() {
	w.out.Debug("Worker.Start linux.")
	w.tapWorker.Start()

	for _, conn := range w.sosWorker {
		conn.Start()
	}

	libol.Go(func() {
		for {
			select {
			case <-w.done:
				return
			case <-w.ticker.C:
				w.SaveStatus()
			}
		}
	})
}

func (w *Worker) Stop() {
	if w.tapWorker == nil || w.conWorker == nil {
		return
	}
	w.done <- true
	w.FreeIpAddr()
	for _, conn := range w.sosWorker {
		conn.Stop()
	}
	w.tapWorker.Stop()
	w.conWorker = nil
	w.tapWorker = nil
}

func (w *Worker) UpTime() int64 {
	client := w.conWorker.client
	if client != nil {
		return client.AliveTime()
	}
	return 0
}

func (w *Worker) FindNext(dest []byte) []byte {
	w.lock.RLock()
	defer w.lock.RUnlock()
	for _, rt := range w.routes {
		if !rt.Destination.Contains(dest) {
			continue
		}
		if w.out.Has(libol.DEBUG) {
			w.out.Debug("Worker.FindNext %v to %v", dest, rt.NextHop)
		} // TODO prefix length
		return rt.NextHop.To4()
	}
	return dest
}

func (w *Worker) OnIpAddr(s *SocketWorker, n *models.Network) error {
	if n.Address == "" {
		w.out.Debug("Worker.OnIpAddr: nil")
		return nil
	}

	addr := fmt.Sprintf("%s/%s", n.Address, n.Netmask)
	if models.NetworkEqual(w.network, n) {
		w.out.Debug("Worker.OnIpAddr: %s noChanged", addr)
		return nil
	}

	w.out.Info("Worker.OnIpAddr: %s %s", addr, n.Routes)

	if w.network == nil {
		prefix := libol.Netmask2Len(n.Netmask)
		ipStr := fmt.Sprintf("%s/%d", n.Address, prefix)
		w.tapWorker.OnIpAddr(ipStr)
		if w.listener.AddAddr != nil {
			_ = w.listener.AddAddr(ipStr, n.Gateway)
		}
	}

	if n.Gateway != "" && runtime.GOOS == "darwin" {
		w.UpdateRoute("0.0.0.0/0", n.Gateway)
	}

	w.network = n
	return nil
}

func (w *Worker) FreeIpAddr() {
	if w.network == nil {
		return
	}
	if w.listener.DelRoutes != nil {
		_ = w.listener.DelRoutes(w.network.Routes)
	}
	if w.listener.DelAddr != nil {
		prefix := libol.Netmask2Len(w.network.Netmask)
		ipStr := fmt.Sprintf("%s/%d", w.network.Address, prefix)
		_ = w.listener.DelAddr(ipStr)
	}
	w.network = nil
	w.routes = make(map[string]PrefixRule)
}

func (w *Worker) OnClose(s *SocketWorker) error {
	w.out.Info("Worker.OnClose")
	w.FreeIpAddr()
	return nil
}

func (w *Worker) OnSuccess(s *SocketWorker) error {
	w.out.Info("Worker.OnSuccess")
	if !w.cfg.RequestAddr {
		w.out.Info("SocketWorker.AddAddr: notAllowed")
	} else if w.listener.AddAddr != nil {
		_ = w.listener.AddAddr(w.ifAddr, "")
	}
	return nil
}

func (w *Worker) UUID() string {
	if w.uuid == "" {
		w.uuid = libol.GenString(13)
	}
	return w.uuid
}

func (w *Worker) SetUUID(v string) {
	w.uuid = v
}

func (w *Worker) UpdateRoute(addr, nexthop string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	dest, _ := libol.ParseCIDR(addr)
	w.routes[dest.String()] = PrefixRule{
		Destination: *dest,
		NextHop:     libol.ParseAddr(nexthop),
	}
}
