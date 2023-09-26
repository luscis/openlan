package _switch

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
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
	bridge    cn.Bridger
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

func (w *OpenLANWorker) updateVPN() {
	cfg := w.Config()
	vCfg := cfg.OpenVPN
	if vCfg == nil {
		return
	}
	routes := vCfg.Routes
	routes = append(routes, vCfg.Subnet)
	if addr := w.Subnet(); addr != "" {
		libol.Info("OpenLANWorker.updateVPN %s subnet %s", cfg.Name, addr)
		routes = append(routes, addr)
	}
	for _, rt := range cfg.Routes {
		addr := rt.Prefix
		if addr == "0.0.0.0/0" {
			vCfg.Push = append(vCfg.Push, "redirect-gateway def1")
			continue
		}
		if _, inet, err := net.ParseCIDR(addr); err == nil {
			routes = append(routes, inet.String())
		}
	}
	vCfg.Routes = routes
}

func (w *OpenLANWorker) forwardVPN() {
	cfg := w.Config()
	vCfg := cfg.OpenVPN
	if vCfg == nil {
		return
	}

	_, port := libol.GetHostPort(vCfg.Listen)
	if vCfg.Protocol == "udp" {
		w.openPort("udp", port, "Open VPN")
	} else {
		w.openPort("tcp", port, "Open VPN")
	}

	devName := vCfg.Device
	// Enable MASQUERADE, and FORWARD it.
	w.toRelated(devName, "Accept related")
	w.toACL(cfg.Acl, devName)

	for _, rt := range vCfg.Routes {
		if rt == "0.0.0.0/0" {
			w.setV.Add("0.0.0.0/1")
			w.setV.Add("128.0.0.0/1")
			break
		}
		w.setV.Add(rt)
	}
	w.toForward_r(devName, vCfg.Subnet, w.setV.Name, "From VPN")
	w.toMasq_r(vCfg.Subnet, w.setV.Name, "From VPN")
}

func (w *OpenLANWorker) forwardSubnet() {
	cfg := w.Config()
	br := cfg.Bridge
	ifAddr := strings.SplitN(br.Address, "/", 2)[0]
	if ifAddr == "" {
		return
	}

	vCfg := w.cfg.OpenVPN
	subnet := w.Subnet()

	// Enable MASQUERADE, and FORWARD it.
	w.toRelated(br.Name, "Accept related")
	for _, rt := range cfg.Routes {
		if rt.MultiPath != nil {
			continue
		}
		if rt.Prefix == "0.0.0.0/0" {
			w.setR.Add("0.0.0.0/1")
			w.setR.Add("128.0.0.0/1")
			break
		}
		w.setR.Add(rt.Prefix)
	}
	w.toForward_r(br.Name, subnet, w.setR.Name, "To route")
	if vCfg != nil {
		w.toMasq_s(w.setR.Name, vCfg.Subnet, "To VPN")
	}
	w.toMasq_r(subnet, w.setR.Name, "To Masq")
}

func (w *OpenLANWorker) Initialize() {
	brCfg := w.cfg.Bridge
	n := models.Network{
		Name:    w.cfg.Name,
		IpStart: w.cfg.Subnet.Start,
		IpEnd:   w.cfg.Subnet.End,
		Netmask: w.cfg.Subnet.Netmask,
		IfAddr:  w.cfg.Bridge.Address,
		Routes:  make([]*models.Route, 0, 2),
	}

	for _, rt := range w.cfg.Routes {
		if rt.NextHop == "" {
			w.out.Warn("OpenLANWorker.Initialize: %s noNextHop", rt.Prefix)
			continue
		}
		rte := models.NewRoute(rt.Prefix, w.IfAddr(), rt.Mode)
		if rt.Metric > 0 {
			rte.Metric = rt.Metric
		}
		if rt.NextHop != "" {
			rte.Origin = rt.NextHop
		}
		n.Routes = append(n.Routes, rte)
	}
	cache.Network.Add(&n)
	for _, ht := range w.cfg.Hosts {
		lease := cache.Network.AddLease(ht.Hostname, ht.Address, n.Name)
		if lease != nil {
			lease.Type = "static"
			lease.Network = w.cfg.Name
		}
	}
	w.bridge = cn.NewBridger(brCfg.Provider, brCfg.Name, brCfg.IPMtu)

	w.updateVPN()
	vCfg := w.cfg.OpenVPN
	if !(vCfg == nil) {
		obj := NewOpenVPN(vCfg)
		obj.Initialize()
		w.vpn = obj
	}
	w.WorkerImpl.Initialize()
	w.forwardSubnet()
	w.forwardVPN()
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

func (w *OpenLANWorker) LoadRoutes() {
	// install routes
	cfg := w.cfg
	w.out.Debug("OpenLANWorker.LoadRoute: %v", cfg.Routes)
	ifAddr := w.IfAddr()
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		if ifAddr == rt.NextHop && rt.MultiPath == nil {
			// route's next-hop is local not install again.
			continue
		}
		nlrt := nl.Route{Dst: dst}
		for _, hop := range rt.MultiPath {
			nxhe := &nl.NexthopInfo{
				Hops: hop.Weight,
				Gw:   net.ParseIP(hop.NextHop),
			}
			nlrt.MultiPath = append(nlrt.MultiPath, nxhe)
		}
		if rt.MultiPath == nil {
			nlrt.Gw = net.ParseIP(rt.NextHop)
			nlrt.Priority = rt.Metric
		}
		w.out.Debug("OpenLANWorker.LoadRoute: %s", nlrt.String())
		promise := &libol.Promise{
			First:  time.Second * 2,
			MaxInt: time.Minute,
			MinInt: time.Second * 10,
		}
		rt_c := rt
		promise.Go(func() error {
			if err := nl.RouteReplace(&nlrt); err != nil {
				w.out.Warn("OpenLANWorker.LoadRoute: %v %s", nlrt, err)
				return err
			}
			w.out.Info("OpenLANWorker.LoadRoute: %v success", rt_c.String())
			return nil
		})
	}
}

func (w *OpenLANWorker) UnLoadRoutes() {
	cfg := w.cfg
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nlRt := nl.Route{Dst: dst}
		if rt.MultiPath == nil {
			nlRt.Gw = net.ParseIP(rt.NextHop)
			nlRt.Priority = rt.Metric
		}
		w.out.Debug("OpenLANWorker.UnLoadRoute: %s", nlRt.String())
		if err := nl.RouteDel(&nlRt); err != nil {
			w.out.Warn("OpenLANWorker.UnLoadRoute: %s", err)
			continue
		}
		w.out.Info("OpenLANWorker.UnLoadRoute: %v", rt.String())
	}
}

func (w *OpenLANWorker) UpBridge(cfg *co.Bridge) {
	master := w.bridge
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
	call := 1
	if w.cfg.Acl == "" {
		call = 0
	}
	if err := master.CallIptables(call); err != nil {
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
	w.LoadRoutes()
	w.WorkerImpl.Start(v)
}

func (w *OpenLANWorker) downBridge(cfg *co.Bridge) {
	w.closePeer(cfg)
	_ = w.bridge.Close()
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
	w.UnLoadRoutes()
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

func (w *OpenLANWorker) Subnet() string {
	cfg := w.cfg

	ipAddr := cfg.Bridge.Address
	ipMask := cfg.Subnet.Netmask
	if ipAddr == "" {
		ipAddr = cfg.Subnet.Start
	}
	if ipAddr != "" {
		addr := ipAddr
		if ipMask != "" {
			prefix := libol.Netmask2Len(ipMask)
			ifAddr := strings.SplitN(ipAddr, "/", 2)[0]
			addr = fmt.Sprintf("%s/%d", ifAddr, prefix)
		}
		if _, inet, err := net.ParseCIDR(addr); err == nil {
			return inet.String()
		}
	}
	return ""
}

func (w *OpenLANWorker) Bridge() cn.Bridger {
	return w.bridge
}

func (w *OpenLANWorker) IfAddr() string {
	return strings.SplitN(w.cfg.Bridge.Address, "/", 2)[0]
}

func (w *OpenLANWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
