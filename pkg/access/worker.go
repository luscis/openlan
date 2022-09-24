package access

import (
	"crypto/tls"
	"fmt"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
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
	AddAddr   func(ipStr string) error
	DelAddr   func(ipStr string) error
	OnTap     func(w *TapWorker) error
	AddRoutes func(routes []*models.Route) error
	DelRoutes func(routes []*models.Route) error
}

type PrefixRule struct {
	Type        int
	Destination net.IPNet
	NextHop     net.IP
}

func GetSocketClient(p *config.Point) libol.SocketClient {
	switch p.Protocol {
	case "kcp":
		c := libol.NewKcpConfig()
		c.Block = config.GetBlock(p.Crypt)
		c.RdQus = p.Queue.SockRd
		c.WrQus = p.Queue.SockWr
		return libol.NewKcpClient(p.Connection, c)
	case "tcp":
		c := &libol.TcpConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		return libol.NewTcpClient(p.Connection, c)
	case "udp":
		c := &libol.UdpConfig{
			Block:   config.GetBlock(p.Crypt),
			Timeout: time.Duration(p.Timeout) * time.Second,
			RdQus:   p.Queue.SockRd,
			WrQus:   p.Queue.SockWr,
		}
		return libol.NewUdpClient(p.Connection, c)
	case "ws":
		c := &libol.WebConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		return libol.NewWebClient(p.Connection, c)
	case "wss":
		c := &libol.WebConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		if p.Cert != nil {
			c.Cert = &libol.CertConfig{
				Insecure: p.Cert.Insecure,
				RootCa:   p.Cert.CaFile,
			}
		}
		return libol.NewWebClient(p.Connection, c)
	default:
		c := &libol.TcpConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		if p.Cert != nil {
			c.Tls = &tls.Config{
				InsecureSkipVerify: p.Cert.Insecure,
				RootCAs:            p.Cert.GetCertPool(),
			}
		}
		return libol.NewTcpClient(p.Connection, c)
	}
}

func GetTapCfg(c *config.Point) network.TapConfig {
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
	tapWorker *TapWorker
	cfg       *config.Point
	uuid      string
	network   *models.Network
	routes    []PrefixRule
	out       *libol.SubLogger
	done      chan bool
	ticker    *time.Ticker
}

func NewWorker(cfg *config.Point) *Worker {
	return &Worker{
		ifAddr: cfg.Interface.Address,
		cfg:    cfg,
		routes: make([]PrefixRule, 0, 32),
		out:    libol.NewSubLogger(cfg.Id()),
		done:   make(chan bool),
		ticker: time.NewTicker(2 * time.Second),
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
	client := GetSocketClient(w.cfg)
	w.conWorker = NewSocketWorker(client, w.cfg)

	tapCfg := GetTapCfg(w.cfg)
	// register listener
	w.tapWorker = NewTapWorker(tapCfg, w.cfg)

	w.conWorker.SetUUID(w.UUID())
	w.conWorker.listener = SocketWorkerListener{
		OnClose:   w.OnClose,
		OnSuccess: w.OnSuccess,
		OnIpAddr:  w.OnIpAddr,
		ReadAt:    w.tapWorker.Write,
	}
	w.conWorker.Initialize()

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
		ReadAt:   w.conWorker.Write,
		FindNext: w.FindNext,
	}
	w.tapWorker.Initialize()
}

func (w *Worker) FlushStatus() {
	file := w.cfg.StatusFile
	device := w.tapWorker.device
	client := w.conWorker.client
	if file == "" || device == nil || client == nil {
		return
	}
	sts := client.Statistics()
	status := &schema.Point{
		RxBytes:   sts[libol.CsRecvOkay],
		TxBytes:   sts[libol.CsSendOkay],
		ErrPkt:    sts[libol.CsSendError],
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
	}
	if w.network != nil {
		status.Address = models.NewNetworkSchema(w.network)
	}
	_ = libol.MarshalSave(status, file, true)
}

func (w *Worker) Start() {
	w.out.Debug("Worker.Start linux.")
	w.FlushStatus()
	w.tapWorker.Start()
	w.conWorker.Start()
	libol.Go(func() {
		for {
			select {
			case <-w.done:
				return
			case <-w.ticker.C:
				w.FlushStatus()
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
	w.conWorker.Stop()
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
	for _, rt := range w.routes {
		if !rt.Destination.Contains(dest) {
			continue
		}
		if rt.Type == 0x00 {
			break
		}
		if w.out.Has(libol.DEBUG) {
			w.out.Debug("Worker.FindNext %v to %v", dest, rt.NextHop)
		}
		return rt.NextHop.To4()
	}
	return dest
}

func (w *Worker) OnIpAddr(s *SocketWorker, n *models.Network) error {
	addr := fmt.Sprintf("%s/%s", n.IfAddr, n.Netmask)
	if models.NetworkEqual(w.network, n) {
		w.out.Debug("Worker.OnIpAddr: %s noChanged", addr)
		return nil
	}
	w.out.Cmd("Worker.OnIpAddr: %s", addr)
	w.out.Cmd("Worker.OnIpAddr: %s", n.Routes)
	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)
	w.tapWorker.OnIpAddr(ipStr)
	if w.listener.AddAddr != nil {
		_ = w.listener.AddAddr(ipStr)
	}
	// Filter routes.
	var routes []*models.Route
	for _, rt := range n.Routes {
		_, _, err := net.ParseCIDR(rt.Prefix)
		if err != nil || rt.NextHop == n.IfAddr {
			continue
		}
		routes = append(routes, rt)
	}
	if w.listener.AddRoutes != nil {
		_ = w.listener.AddRoutes(routes)
	}
	w.network = n
	// update routes
	ip := net.ParseIP(w.network.IfAddr)
	m := net.IPMask(net.ParseIP(w.network.Netmask).To4())
	w.routes = append(w.routes, PrefixRule{
		Type:        0x00,
		Destination: net.IPNet{IP: ip.Mask(m), Mask: m},
		NextHop:     libol.EthZero,
	})
	for _, rt := range routes {
		_, dest, _ := net.ParseCIDR(rt.Prefix)
		w.routes = append(w.routes, PrefixRule{
			Type:        0x01,
			Destination: *dest,
			NextHop:     net.ParseIP(rt.NextHop),
		})
	}
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
		ipStr := fmt.Sprintf("%s/%d", w.network.IfAddr, prefix)
		_ = w.listener.DelAddr(ipStr)
	}
	w.network = nil
	w.routes = make([]PrefixRule, 0, 32)
}

func (w *Worker) OnClose(s *SocketWorker) error {
	w.out.Info("Worker.OnClose")
	w.FreeIpAddr()
	return nil
}

func (w *Worker) OnSuccess(s *SocketWorker) error {
	w.out.Info("Worker.OnSuccess")
	if w.listener.AddAddr != nil {
		_ = w.listener.AddAddr(w.ifAddr)
	}
	return nil
}

func (w *Worker) UUID() string {
	if w.uuid == "" {
		w.uuid = libol.GenRandom(13)
	}
	return w.uuid
}

func (w *Worker) SetUUID(v string) {
	w.uuid = v
}
