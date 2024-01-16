package cswitch

import (
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	nl "github.com/vishvananda/netlink"
)

func PeerName(name, prefix string) (string, string) {
	return name + prefix + "i", name + prefix + "o"
}

type OpenLANWorker struct {
	*WorkerImpl
	alias     string
	newTime   int64
	startTime int64
	links     *Links
}

func NewOpenLANWorker(c *co.Network) *OpenLANWorker {
	return &OpenLANWorker{
		WorkerImpl: NewWorkerApi(c),
		alias:      c.Alias,
		newTime:    time.Now().Unix(),
		startTime:  0,
		links:      NewLinks(),
	}
}

func (w *OpenLANWorker) Initialize() {
	brCfg := w.cfg.Bridge
	name := w.cfg.Name

	w.br = cn.NewBridger(brCfg.Provider, brCfg.Name, brCfg.IPMtu)

	for _, ht := range w.cfg.Hosts {
		lease := cache.Network.AddLease(ht.Hostname, ht.Address, name)
		if lease != nil {
			lease.Type = "static"
			lease.Network = name
		}
	}

	w.WorkerImpl.Initialize()
}

func (w *OpenLANWorker) LoadLinks() {
	if w.cfg.Links != nil {
		for _, link := range w.cfg.Links {
			link.Correct()
			w.AddLink(link)
		}
	}
}

func (w *OpenLANWorker) UnLoadLinks() {
	w.links.lock.RLock()
	defer w.links.lock.RUnlock()
	for _, l := range w.links.links {
		l.Stop()
	}
}

func (w *OpenLANWorker) UpBridge(cfg *co.Bridge) {
	master := w.br
	// new it and configure address
	master.Open(cfg.Address)
	// configure stp
	if cfg.Stp == "enable" {
		if err := master.Stp(true); err != nil {
			w.out.Warn("OpenLANWorker.UpBridge: Stp %s", err)
		}
	} else {
		_ = master.Stp(false)
	}
	// configure forward delay
	if err := master.Delay(cfg.Delay); err != nil {
		w.out.Warn("OpenLANWorker.UpBridge: Delay %s", err)
	}
	w.connectPeer(cfg)
	if err := master.CallIptables(0); err != nil {
		w.out.Warn("OpenLANWorker.Start: CallIptables %s", err)
	}
}

func (w *OpenLANWorker) connectPeer(cfg *co.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in, ex := PeerName(cfg.Network, "-e")
	link := &nl.Veth{
		LinkAttrs: nl.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	br := cn.NewBrCtl(cfg.Peer, cfg.IPMtu)
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		if !br.Has() {
			w.out.Warn("%s notFound", br.Name)
			return libol.NewErr("%s notFound", br.Name)
		}
		err := nl.LinkAdd(link)
		if err != nil {
			w.out.Error("OpenLANWorker.connectPeer: %s", err)
			return nil
		}
		br0 := cn.NewBrCtl(cfg.Name, cfg.IPMtu)
		if err := br0.AddPort(in); err != nil {
			w.out.Error("OpenLANWorker.connectPeer: %s", err)
		}
		br1 := cn.NewBrCtl(cfg.Peer, cfg.IPMtu)
		if err := br1.AddPort(ex); err != nil {
			w.out.Error("OpenLANWorker.connectPeer: %s", err)
		}
		return nil
	})
}

func (w *OpenLANWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.startTime = time.Now().Unix()

	w.out.Info("OpenLANWorker.Start")

	w.UpBridge(w.cfg.Bridge)
	w.LoadLinks()

	w.WorkerImpl.Start(v)
}

func (w *OpenLANWorker) downBridge(cfg *co.Bridge) {
	w.closePeer(cfg)
	_ = w.br.Close()
}

func (w *OpenLANWorker) closePeer(cfg *co.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in, ex := PeerName(cfg.Network, "-e")
	link := &nl.Veth{
		LinkAttrs: nl.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	err := nl.LinkDel(link)
	if err != nil {
		w.out.Error("OpenLANWorker.closePeer: %s", err)
		return
	}
}

func (w *OpenLANWorker) Stop() {
	w.out.Info("OpenLANWorker.Close")
	w.WorkerImpl.Stop()
	w.UnLoadLinks()
	w.startTime = 0
	w.downBridge(w.cfg.Bridge)
}

func (w *OpenLANWorker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *OpenLANWorker) AddLink(c co.Point) {
	br := w.cfg.Bridge
	uuid := libol.GenString(13)

	c.Alias = w.alias
	c.Network = w.cfg.Name
	c.RequestAddr = false
	c.Interface.Name = cn.Taps.GenName()
	c.Interface.Bridge = br.Name
	c.Interface.Address = br.Address
	c.Interface.Provider = br.Provider
	c.Interface.IPMtu = br.IPMtu
	c.Log.File = "/dev/null"

	l := NewLink(uuid, &c)
	l.Initialize()
	cache.Link.Add(uuid, l.Model())
	w.links.Add(l)
	l.Start()
}

func (w *OpenLANWorker) DelLink(addr string) {
	if l := w.links.Remove(addr); l != nil {
		cache.Link.Del(l.uuid)
	}
}

func (w *OpenLANWorker) Bridge() cn.Bridger {
	return w.br
}

func (w *OpenLANWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
