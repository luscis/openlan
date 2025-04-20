package cswitch

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

func NewNetworker(c *co.Network) api.Networker {
	var obj api.Networker
	switch c.Provider {
	case "ipsec":
		secer := NewIPSecWorker(c)
		api.Call.SetIPSecer(secer)
		obj = secer
	case "router":
		obj = NewRouterWorker(c)
	default:
		obj = NewOpenLANWorker(c)
	}
	api.Call.AddWorker(c.Name, obj)
	return obj
}

func SplitCombined(value string) (string, string) {
	values := strings.SplitN(value, ":", 2)
	if len(values) == 2 {
		return values[0], values[1]
	}
	return values[0], ""
}

type WorkerImpl struct {
	uuid    string
	cfg     *co.Network
	out     *libol.SubLogger
	dhcp    *Dhcp
	fire    *cn.FireWallTable
	setR    *cn.IPSet
	setV    *cn.IPSet
	vpn     *OpenVPN
	ztrust  *ZTrust
	qos     *QosCtrl
	vrf     *cn.VRF
	table   int
	br      cn.Bridger
	acl     *ACL
	findhop *FindHop
	snat    *cn.FireWallChain
}

func NewWorkerApi(c *co.Network) *WorkerImpl {
	return &WorkerImpl{
		cfg:   c,
		out:   libol.NewSubLogger(c.Name),
		setR:  cn.NewIPSet(c.Name+"_r", "hash:net"),
		setV:  cn.NewIPSet(c.Name+"_v", "hash:net"),
		table: 0,
	}
}

func (w *WorkerImpl) Provider() string {
	return w.cfg.Provider
}

func (w *WorkerImpl) newRoute(rt *co.PrefixRoute) *models.Route {
	if rt.NextHop == "" {
		w.out.Warn("WorkerImpl.NewRoute: %s noNextHop", rt.Prefix)
		return nil
	}
	rte := models.NewRoute(rt.Prefix, w.IfAddr())
	if rt.Metric > 0 {
		rte.Metric = rt.Metric
	}
	if rt.NextHop != "" {
		rte.Origin = rt.NextHop
	}
	return rte
}

func (w *WorkerImpl) Initialize() {
	cfg := w.cfg

	if cfg.Namespace != "" {
		w.vrf = cn.NewVRF(cfg.Namespace, 0)
		w.table = w.vrf.Table()
	}

	w.acl = NewACL(cfg.Name)
	w.acl.Initialize()

	w.findhop = NewFindHop(cfg.Name, cfg)
	if cfg.Subnet != nil {
		n := models.Network{
			Name:    cfg.Name,
			IpStart: cfg.Subnet.Start,
			IpEnd:   cfg.Subnet.End,
			Netmask: cfg.Subnet.Netmask,
			Address: cfg.Bridge.Address,
			Config:  cfg,
		}
		cache.Network.Add(&n)
	}

	w.updateVPN()
	w.createVPN()

	w.fire = cn.NewFireWallTable(cfg.Name)
	w.snat = cn.NewFireWallChain("XTT_"+cfg.Name+"_SNAT", cn.TNat, "")

	if out, err := w.setV.Clear(); err != nil {
		w.out.Error("WorkerImpl.Initialize: create ipset: %s %s", out, err)
	}
	if out, err := w.setR.Clear(); err != nil {
		w.out.Error("WorkerImpl.Initialize: create ipset: %s %s", out, err)
	}

	w.ztrust = NewZTrust(cfg.Name, 30)
	w.ztrust.Initialize()

	w.qos = NewQosCtrl(cfg.Name)
	w.qos.Initialize()

	if cfg.Dhcp == "enable" {
		name := cfg.Bridge.Name
		if w.br != nil {
			name = w.br.L3Name()
		}
		w.dhcp = NewDhcp(&co.Dhcp{
			Name:      cfg.Name,
			Subnet:    cfg.Subnet,
			Interface: name,
		})
	}

	w.forwardSubnet()
	w.forwardVPN()
}

func (w *WorkerImpl) AddPhysical(bridge string, output string) {
	br := cn.NewBrCtl(bridge, 0)
	if err := br.AddPort(output); err != nil {
		w.out.Warn("WorkerImpl.AddPhysical %s", err)
	}
}

func (w *WorkerImpl) addOutput(bridge string, port *co.Output) {
	mtu := 0
	if port.Protocol == "gre" {
		mtu = 1450
		link := &LinuxLink{
			link: &nl.Gretap{
				IKey: uint32(port.Segment),
				OKey: uint32(port.Segment),
				LinkAttrs: nl.LinkAttrs{
					Name: port.Link,
					MTU:  mtu,
				},
				Local:    libol.ParseAddr("0.0.0.0"),
				Remote:   libol.ParseAddr(port.Remote),
				PMtuDisc: 1,
			},
		}
		if err := link.Start(); err != nil {
			w.out.Error("WorkerImpl.LinkStart %s %s", port.Id(), err)
			return
		}
		port.Linker = link
	} else if port.Protocol == "vxlan" {
		dport := 8472
		if port.DstPort > 0 {
			dport = port.DstPort
		}
		mtu = 1450
		link := &LinuxLink{
			link: &nl.Vxlan{
				LinkAttrs: nl.LinkAttrs{
					Name: port.Link,
				},
			},
		}
		opts := []string{"type", "vxlan",
			"id", strconv.Itoa(port.Segment),
			"remote", port.Remote,
			"dstport", strconv.Itoa(dport),
			"noudpcsum"}
		_, err := cn.LinkAdd(port.Link, opts...)
		if err != nil {
			w.out.Error("WorkerImpl.LinkStart %s %v", port.Id(), opts)
			return
		}
		cn.LinkSet(port.Link, "mtu", strconv.Itoa(mtu))
		port.Linker = link
	} else if port.Protocol == "tcp" || port.Protocol == "tls" ||
		port.Protocol == "wss" {
		port.Link = cn.Taps.GenName()
		name, pass := SplitCombined(port.Secret)
		algo, secret := SplitCombined(port.Crypt)
		ac := co.Point{
			Alias:       w.cfg.Alias,
			Network:     w.cfg.Name,
			RequestAddr: false,
			Interface: co.Interface{
				Name:   port.Link,
				Bridge: bridge,
			},
			Log: co.Log{
				File: "/dev/null",
			},
			Connection: port.Remote,
			Protocol:   port.Protocol,
			Username:   name,
			Password:   pass,
		}
		if secret != "" {
			ac.Crypt = &co.Crypt{
				Algo:   algo,
				Secret: secret,
			}
		}
		link := NewLink(&ac)
		link.Initialize()
		link.Start()
		port.Linker = link
	} else {
		link, err := nl.LinkByName(port.Remote)
		if link == nil {
			w.out.Error("WorkerImpl.addOutput %s %s", port.Remote, err)
			return
		}
		if err := nl.LinkSetUp(link); err != nil {
			w.out.Warn("WorkerImpl.addOutput %s %s", port.Remote, err)
		}

		if port.Segment > 0 {
			subLink := &LinuxLink{
				link: &nl.Vlan{
					LinkAttrs: nl.LinkAttrs{
						Name:        port.Link,
						ParentIndex: link.Attrs().Index,
					},
					VlanId: port.Segment,
				},
			}
			if err := subLink.Start(); err != nil {
				w.out.Error("WorkerImpl.LinkStart %s %s", port.Link, err)
				return
			}
			port.Linker = subLink
		}
	}

	if mtu > 0 {
		if w.br != nil {
			w.br.SetMtu(mtu)
		}
	}

	out := &models.Output{
		Network:  w.cfg.Name,
		NewTime:  time.Now().Unix(),
		Protocol: port.Protocol,
		Remote:   port.Remote,
		Segment:  port.Segment,
		Device:   port.Link,
		Secret:   port.Secret,
	}
	cache.Output.Add(port.Link, out)

	w.out.Info("WorkerImpl.addOutput %s %s", port.Link, port.Id())
	w.AddPhysical(bridge, port.Link)
}

func (w *WorkerImpl) loadRoute(rt co.PrefixRoute) {
	// install routes
	ifAddr := w.IfAddr()

	dst, err := libol.ParseNet(rt.Prefix)
	if err != nil {
		return
	}
	if ifAddr == rt.NextHop && rt.MultiPath == nil && rt.FindHop == "" {
		// route's next-hop is local not install again.
		return
	}
	nlr := nl.Route{
		Dst:   dst,
		Table: w.table,
	}
	for _, hop := range rt.MultiPath {
		nxhe := &nl.NexthopInfo{
			Hops: hop.Weight,
			Gw:   net.ParseIP(hop.NextHop),
		}
		nlr.MultiPath = append(nlr.MultiPath, nxhe)
	}
	if rt.MultiPath == nil {
		nlr.Gw = net.ParseIP(rt.NextHop)
		nlr.Priority = rt.Metric
	}
	if rt.FindHop != "" {
		w.findhop.LoadHop(rt.FindHop, &nlr)
		return
	}
	w.out.Info("WorkerImpl.loadRoute: %s", nlr.String())

	rt_c := rt
	promise := libol.NewPromise()
	promise.Go(func() error {
		if err := nl.RouteReplace(&nlr); err != nil {
			w.out.Warn("WorkerImpl.loadRoute: %v %s", nlr, err)
			return err
		}
		w.out.Info("WorkerImpl.loadRoute: %v success", rt_c.String())
		return nil
	})
}

func (w *WorkerImpl) loadRoutes() {
	// install routes
	cfg := w.cfg
	w.out.Info("WorkerImpl.LoadRoute: %v", cfg.Routes)

	for _, rt := range cfg.Routes {
		w.loadRoute(rt)
	}
}

func (w *WorkerImpl) loadVRF() {
	if w.vrf == nil {
		return
	}

	if err := w.vrf.Up(); err != nil {
		w.out.Warn("OpenLANWorker.UpVRF %s", err)
		return
	}

	if w.br != nil {
		if err := w.vrf.AddSlave(w.br.L3Name()); err != nil {
			w.out.Warn("OpenLANWorker.UpVRF %s", err)
			return
		}
	}
}

func (w *WorkerImpl) setMss() {
	cfg, _ := w.GetCfgs()

	mss := cfg.Bridge.Mss
	w.fire.Mangle.Post.AddRuleX(cn.IPRule{
		Order:   "-I",
		Output:  cfg.Bridge.Name,
		Proto:   "tcp",
		Match:   "tcp",
		TcpFlag: []string{"SYN,RST", "SYN"},
		Jump:    cn.CTcpMss,
		SetMss:  mss,
	})
	if w.br != nil {
		w.fire.Mangle.Post.AddRuleX(cn.IPRule{
			Order:   "-I",
			Output:  w.br.L3Name(),
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    cn.CTcpMss,
			SetMss:  mss,
		})
	}
	// connect from local
	w.fire.Mangle.In.AddRuleX(cn.IPRule{
		Order:   "-I",
		Input:   cfg.Bridge.Name,
		Proto:   "tcp",
		Match:   "tcp",
		TcpFlag: []string{"SYN,RST", "SYN"},
		Jump:    cn.CTcpMss,
		SetMss:  mss,
	})
}

func (w *WorkerImpl) SetMss(mss int) {
	cfg, _ := w.GetCfgs()
	if cfg.Bridge.Mss != mss {
		cfg.Bridge.Mss = mss
		w.setMss()
	}
}

func (w *WorkerImpl) doSnat() {
	w.fire.Nat.Post.AddRuleX(cn.IPRule{
		Jump:    w.snat.Chain().Name,
		Comment: "Goto SNAT",
	})
}

func (w *WorkerImpl) undoSnat() {
	w.fire.Nat.Post.DelRuleX(cn.IPRule{
		Jump:    w.snat.Chain().Name,
		Comment: "Goto SNAT",
	})
}

func (w *WorkerImpl) DoSnat() {
	cfg, _ := w.GetCfgs()
	if cfg.Snat == "disable" {
		cfg.Snat = "enable"
		w.doSnat()
	}
}

func (w *WorkerImpl) UndoSnat() {
	cfg, _ := w.GetCfgs()
	if cfg.Snat != "disable" {
		cfg.Snat = "disable"
		w.undoSnat()
	}
}

func (w *WorkerImpl) doTrust() {
	_, vpn := w.GetCfgs()
	w.fire.Mangle.Pre.AddRuleX(cn.IPRule{
		Input:   vpn.Device,
		Jump:    w.ztrust.Chain(),
		Comment: "Goto Zero Trust",
	})
}

func (w *WorkerImpl) undoTrust() {
	_, vpn := w.GetCfgs()
	w.fire.Mangle.Pre.DelRuleX(cn.IPRule{
		Input:   vpn.Device,
		Jump:    w.ztrust.Chain(),
		Comment: "Goto Zero Trust",
	})
}

func (w *WorkerImpl) DoZTrust() {
	cfg, _ := w.GetCfgs()
	if cfg.ZTrust != "enable" {
		cfg.ZTrust = "enable"
		w.doTrust()
	}
}

func (w *WorkerImpl) UndoZTrust() {
	cfg, _ := w.GetCfgs()
	if cfg.ZTrust == "enable" {
		cfg.ZTrust = "disable"
		w.undoTrust()
	}
}

func (w *WorkerImpl) letVPN2VRF() {
	_, vpn := w.GetCfgs()
	promise := libol.NewPromise()
	promise.Go(func() error {
		link, err := nl.LinkByName(vpn.Device)
		if link == nil {
			w.out.Info("Link %s %s", vpn.Device, err)
			return err
		}

		attr := link.Attrs()
		if err := w.vrf.AddSlave(attr.Name); err != nil {
			w.out.Info("VRF AddSlave: %s", err)
			return err
		}

		dest, _ := libol.ParseNet(vpn.Subnet)
		rt := &nl.Route{
			Dst:       dest,
			Table:     w.table,
			LinkIndex: attr.Index,
		}
		w.out.Debug("WorkerImpl.LoadRoute: %s", rt.String())
		if err := nl.RouteAdd(rt); err != nil {
			w.out.Warn("Route add: %s", err)
			return err
		}

		return nil
	})
}

func (w *WorkerImpl) Start(v api.Switcher) {
	cfg, vpn := w.GetCfgs()

	w.out.Info("WorkerImpl.Start")

	w.loadVRF()
	w.loadRoutes()

	w.acl.Start()
	w.toACL(cfg.Bridge.Name)

	for _, output := range cfg.Outputs {
		output.GenName()
		w.addOutput(cfg.Bridge.Name, output)
	}

	if !(w.vpn == nil) {
		w.vpn.Start()
		if !(w.vrf == nil) {
			w.letVPN2VRF()
		}
		w.fire.Mangle.In.AddRule(cn.IPRule{
			Input:   vpn.Device,
			Jump:    w.qos.ChainIn(),
			Comment: "Goto Qos ChainIn",
		})
		w.qos.Start()
		w.ztrust.Start()
	}

	w.fire.Start()
	w.snat.Install()
	if cfg.Snat != "disable" {
		w.doSnat()
	}
	if cfg.Bridge.Mss > 0 {
		// forward to remote
		w.setMss()
	}

	w.findhop.Start()

	if !(w.dhcp == nil) {
		w.dhcp.Start()
	}
	if !(w.vpn == nil) {
		if cfg.ZTrust == "enable" {
			w.doTrust()
		}
	}
}

func (w *WorkerImpl) DelPhysical(bridge string, output string) {
	br := cn.NewBrCtl(bridge, 0)
	if err := br.DelPort(output); err != nil {
		w.out.Warn("WorkerImpl.DelPhysical %s", err)
	}
}

func (w *WorkerImpl) delOutput(bridge string, port *co.Output) {
	w.out.Info("WorkerImpl.delOutput %s", port.Link)

	cache.Output.Del(port.Link)
	w.DelPhysical(bridge, port.Link)

	link := port.Linker
	if link != nil {
		if err := link.Stop(); err != nil {
			w.out.Error("WorkerImpl.LinkStop %s %s", port.Link, err)
			return
		}
	}
}

func (w *WorkerImpl) unloadRoute(rt co.PrefixRoute) {
	dst, err := libol.ParseNet(rt.Prefix)
	if err != nil {
		return
	}
	nlr := nl.Route{
		Dst:   dst,
		Table: w.table,
	}
	if rt.MultiPath == nil {
		nlr.Gw = net.ParseIP(rt.NextHop)
		nlr.Priority = rt.Metric
	}

	if rt.FindHop != "" {
		w.findhop.UnloadHop(rt.FindHop, &nlr)
		return
	}
	w.out.Debug("WorkerImpl.UnLoadRoute: %s", nlr.String())
	if err := nl.RouteDel(&nlr); err != nil {
		w.out.Warn("WorkerImpl.UnLoadRoute: %s", err)
		return
	}
	w.out.Info("WorkerImpl.UnLoadRoute: %v", rt.String())
}

func (w *WorkerImpl) unloadRoutes() {
	cfg := w.cfg
	for _, rt := range cfg.Routes {
		w.unloadRoute(rt)
	}
}

func (w *WorkerImpl) RestartVPN() {
	if w.vpn != nil {
		w.vpn.Restart()
		if !(w.vrf == nil) {
			w.letVPN2VRF()
		}
	}
}

func (w *WorkerImpl) Stop() {
	w.out.Info("WorkerImpl.Stop")

	cfg, _ := w.GetCfgs()
	if cfg.Snat != "disable" {
		w.undoSnat()
	}

	w.snat.Cancel()
	w.fire.Stop()
	w.findhop.Stop()
	w.acl.Stop()

	w.unloadRoutes()

	if !(w.vpn == nil) {
		w.ztrust.Stop()
		w.qos.Stop()
		w.vpn.Stop()
	}
	if !(w.dhcp == nil) {
		w.dhcp.Stop()
	}
	if !(w.vrf == nil) {
		w.vrf.Down()
	}

	for _, output := range cfg.Outputs {
		w.delOutput(cfg.Bridge.Name, output)
	}

	w.setR.Destroy()
	w.setV.Destroy()
}

func (w *WorkerImpl) String() string {
	return w.cfg.Name
}

func (w *WorkerImpl) ID() string {
	return w.uuid
}

func (w *WorkerImpl) Bridger() cn.Bridger {
	return w.br
}

func (w *WorkerImpl) Config() *co.Network {
	return w.cfg
}

func (w *WorkerImpl) Subnet() *net.IPNet {
	cfg := w.cfg

	ipAddr := cfg.Bridge.Address
	ipMask := cfg.Subnet.Netmask
	if ipAddr == "" {
		ipAddr = cfg.Subnet.Start
	}
	if ipAddr == "" {
		return nil
	}

	addr := ipAddr
	if ipMask != "" {
		prefix := libol.Netmask2Len(ipMask)
		ifAddr := strings.SplitN(ipAddr, "/", 2)[0]
		addr = fmt.Sprintf("%s/%d", ifAddr, prefix)
	}
	if inet, err := libol.ParseNet(addr); err == nil {
		return inet
	}

	return nil
}

func (w *WorkerImpl) Reload(v api.Switcher) {
}

func (w *WorkerImpl) toACL(input string) {
	if input == "" {
		return
	}
	w.fire.Raw.Pre.AddRule(cn.IPRule{
		Input: input,
		Jump:  w.acl.Chain(),
	})
}

func (w *WorkerImpl) openPort(protocol, port, comment string) {
	w.out.Info("WorkerImpl.openPort %s %s", protocol, port)
	// allowed forward between source and prefix.
	w.fire.Filter.In.AddRule(cn.IPRule{
		Proto:   protocol,
		Match:   "multiport",
		DstPort: port,
		Comment: comment,
	})
}

func (w *WorkerImpl) toForward_i(input, pfxSet, comment string) {
	w.out.Debug("WorkerImpl.toForward %s %s:%s", input, pfxSet)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.AddRule(cn.IPRule{
		Input:   input,
		DestSet: pfxSet,
		Comment: comment,
	})
}

func (w *WorkerImpl) toForward_r(input, source, pfxSet, comment string) {
	w.out.Debug("WorkerImpl.toForward %s:%s %s:%s", input, source, pfxSet)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.AddRule(cn.IPRule{
		Input:   input,
		Source:  source,
		DestSet: pfxSet,
		Comment: comment,
	})
}

func (w *WorkerImpl) toForward_s(input, srcSet, prefix, comment string) {
	w.out.Debug("WorkerImpl.toForward %s:%s %s:%s", input, srcSet, prefix)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.AddRule(cn.IPRule{
		Input:   input,
		SrcSet:  srcSet,
		Dest:    prefix,
		Comment: comment,
	})
}

func (w *WorkerImpl) toMasq_r(source, pfxSet, comment string) {
	// Enable masquerade from source to prefix.
	output := ""
	w.snat.AddRule(cn.IPRule{
		Mark:    uint32(w.table),
		Source:  source,
		DestSet: pfxSet,
		Output:  output,
		Jump:    cn.CMasq,
		Comment: comment,
	})

}

func (w *WorkerImpl) toMasq_i(input, pfxSet, comment string) {
	// Enable masquerade from input to prefix.
	w.snat.AddRule(cn.IPRule{
		Mark:    uint32(w.table),
		Input:   input,
		DestSet: pfxSet,
		Jump:    cn.CMasq,
		Comment: comment,
	})

}

func (w *WorkerImpl) toMasq_s(srcSet, prefix, comment string) {
	// Enable masquerade from source to prefix.
	w.snat.AddRule(cn.IPRule{
		Mark:    uint32(w.table),
		SrcSet:  srcSet,
		Dest:    prefix,
		Jump:    cn.CMasq,
		Comment: comment,
	})

}

func (w *WorkerImpl) toRelated(output, comment string) {
	w.out.Debug("WorkerImpl.toRelated %s", output)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.AddRule(cn.IPRule{
		Output:  output,
		CtState: "RELATED,ESTABLISHED",
		Comment: comment,
	})
}

func (w *WorkerImpl) GetCfgs() (*co.Network, *co.OpenVPN) {
	cfg := w.cfg
	vpn := cfg.OpenVPN
	return cfg, vpn
}

func (w *WorkerImpl) updateVPNRoute(routes []string, rt co.PrefixRoute) []string {
	_, vpn := w.GetCfgs()
	if vpn == nil {
		return routes
	}

	addr := rt.Prefix
	if addr == "0.0.0.0/0" {
		vpn.AddRedirectDef1()
		routes = append(routes, addr)
		return routes
	}
	if inet, err := libol.ParseNet(addr); err == nil {
		routes = append(routes, inet.String())
	}

	return routes
}

func (w *WorkerImpl) updateVPN() {
	cfg, vpn := w.GetCfgs()
	if vpn == nil {
		return
	}

	routes := vpn.Routes
	routes = append(routes, vpn.Subnet) // add subnet of VPN self.
	if addr := w.Subnet(); addr != nil {
		w.out.Info("WorkerImpl.updateVPN subnet %s", addr)
		routes = append(routes, addr.String())
	}

	for _, rt := range cfg.Routes {
		routes = w.updateVPNRoute(routes, rt)
	}
	vpn.Routes = routes
}

func (w *WorkerImpl) forwardZone(input string) {
	if w.table == 0 {
		return
	}

	w.out.Debug("WorkerImpl.forwardZone %s", input)
	w.fire.Raw.Pre.AddRule(cn.IPRule{
		Input:   input,
		Jump:    cn.CMark,
		SetMark: uint32(w.table),
		Comment: "Mark private traffic",
	})
	w.fire.Raw.Pre.AddRule(cn.IPRule{
		Input:   input,
		Jump:    cn.CCT,
		Zone:    uint32(w.table),
		Comment: "Goto private zone",
	})
	w.fire.Raw.Out.AddRule(cn.IPRule{
		Output:  input,
		Jump:    cn.CCT,
		Zone:    uint32(w.table),
		Comment: "Goto private zone",
	})
}

func (w *WorkerImpl) addVPNSet(rt string) {
	if rt == "0.0.0.0/0" {
		w.setV.Add("0.0.0.0/1")
		w.setV.Add("128.0.0.0/1")
		return
	}
	w.setV.Add(rt)
}

func (w *WorkerImpl) delVPNSet(rt string) {
	if rt == "0.0.0.0/0" {
		w.setV.Del("0.0.0.0/1")
		w.setV.Del("128.0.0.0/1")
		return
	}
	w.setV.Del(rt)
}

func (w *WorkerImpl) forwardVPN() {
	_, vpn := w.GetCfgs()
	if vpn == nil {
		return
	}

	devName := vpn.Device
	_, port := libol.GetHostPort(vpn.Listen)
	if vpn.Protocol == "udp" {
		w.openPort("udp", port, "Open VPN")
	} else {
		w.openPort("tcp", port, "Open VPN")
	}

	w.forwardZone(devName)

	// Enable MASQUERADE, and FORWARD it.
	w.toRelated(devName, "Accept related")
	w.toACL(devName)

	for _, rt := range vpn.Routes {
		w.addVPNSet(rt)
	}

	if w.vrf != nil {
		w.toForward_r(w.vrf.Name(), vpn.Subnet, w.setV.Name, "From VPN")
	} else {
		w.toForward_r(devName, vpn.Subnet, w.setV.Name, "From VPN")
	}
	w.toMasq_r(vpn.Subnet, w.setV.Name, "From VPN")
}

func (w *WorkerImpl) addIpSet(rt co.PrefixRoute) bool {
	if rt.MultiPath != nil {
		return true
	}
	if rt.Prefix == "0.0.0.0/0" {
		w.setR.Add("0.0.0.0/1")
		w.setR.Add("128.0.0.0/1")
		return false
	}
	w.setR.Add(rt.Prefix)

	return true
}

func (w *WorkerImpl) delIpSet(rt co.PrefixRoute) {
	if rt.MultiPath != nil {
		return
	}
	if rt.Prefix == "0.0.0.0/0" {
		w.setR.Del("0.0.0.0/1")
		w.setR.Del("128.0.0.0/1")
		return
	}
	w.setR.Del(rt.Prefix)
	return
}

func (w *WorkerImpl) forwardSubnet() {
	cfg, vpn := w.GetCfgs()

	input := cfg.Bridge.Name
	if w.br != nil {
		input = w.br.L3Name()
		w.forwardZone(input)
	}

	ifAddr := strings.SplitN(cfg.Bridge.Address, "/", 2)[0]
	if ifAddr == "" {
		return
	}

	// Enable MASQUERADE, and FORWARD it.
	w.toRelated(input, "Accept related")
	for _, rt := range cfg.Routes {
		if !w.addIpSet(rt) {
			break
		}
	}

	if w.vrf != nil {
		w.toForward_i(w.vrf.Name(), w.setR.Name, "To route")
	} else {
		w.toForward_i(input, w.setR.Name, "To route")
	}

	if vpn != nil {
		w.toMasq_s(w.setR.Name, vpn.Subnet, "To VPN")
	}
	w.toMasq_i(input, w.setR.Name, "To Masq")
}

func (w *WorkerImpl) createVPN() {
	_, vpn := w.GetCfgs()
	if vpn == nil {
		return
	}

	obj := NewOpenVPN(vpn)
	obj.Initialize()
	w.vpn = obj
}

func (w *WorkerImpl) addVPNRoute(rt co.PrefixRoute) {
	vpn := w.cfg.OpenVPN
	if vpn == nil {
		return
	}

	routes := vpn.Routes
	vpn.Routes = w.updateVPNRoute(routes, rt)
}

func (w *WorkerImpl) delVPNRoute(rt co.PrefixRoute) {
	vpn := w.cfg.OpenVPN
	if vpn == nil {
		return
	}

	routes := vpn.Routes
	addr := rt.Prefix
	if addr == "0.0.0.0/0" {
		vpn.DelRedirectDef1()
		for i2, r := range routes {
			if r == addr {
				routes = append(routes[:i2], routes[i2+1:]...)
				break
			}
		}
		return
	}
	if inet, err := libol.ParseNet(addr); err == nil {
		for i, r := range routes {
			if r == inet.String() {
				routes = append(routes[:i], routes[i+1:]...)
				break
			}
		}
	}

	vpn.Routes = routes
}

func (w *WorkerImpl) correctRoute(route *schema.PrefixRoute) co.PrefixRoute {
	rt := co.PrefixRoute{
		Prefix:  route.Prefix,
		NextHop: route.NextHop,
		FindHop: route.FindHop,
		Metric:  route.Metric,
	}
	rt.CorrectRoute(w.IfAddr())
	return rt
}

func (w *WorkerImpl) ListRoute(call func(obj schema.PrefixRoute)) {
	w.cfg.ListRoute(func(obj co.PrefixRoute) {
		data := schema.PrefixRoute{
			Prefix:  obj.Prefix,
			NextHop: obj.NextHop,
			FindHop: obj.FindHop,
			Metric:  obj.Metric,
		}
		call(data)
	})
}

func (w *WorkerImpl) AddRoute(route *schema.PrefixRoute, switcher api.Switcher) error {
	rt := w.correctRoute(route)
	if !w.cfg.AddRoute(rt) {
		w.out.Info("WorkerImpl.AddRoute: %s route exist", route.Prefix)
		return nil
	}

	w.out.Info("WorkerImpl.AddRoute: %s", rt.String())
	w.addIpSet(rt)
	if inet, err := libol.ParseNet(rt.Prefix); err == nil {
		w.addVPNSet(inet.String())
	}
	w.addVPNRoute(rt)
	w.loadRoute(rt)
	return nil
}

func (w *WorkerImpl) DelRoute(route *schema.PrefixRoute, switcher api.Switcher) error {
	correctRt := w.correctRoute(route)
	delRt, removed := w.cfg.DelRoute(correctRt)
	if !removed {
		w.out.Info("WorkerImpl.DelRoute: %s not found", route.Prefix)
		return nil
	}

	w.delIpSet(delRt)
	if inet, err := libol.ParseNet(delRt.Prefix); err == nil {
		w.delVPNSet(inet.String())
	}
	w.delVPNRoute(delRt)
	w.unloadRoute(delRt)
	return nil
}

func (w *WorkerImpl) SaveRoute() {
	w.cfg.SaveRoute()
}

func (w *WorkerImpl) Router() api.Router {
	return w
}

func (w *WorkerImpl) IfAddr() string {
	br := w.cfg.Bridge
	return strings.SplitN(br.Address, "/", 2)[0]
}

func (w *WorkerImpl) AddOutput(data schema.Output) {
	output := &co.Output{
		Segment:  data.Segment,
		Protocol: data.Protocol,
		Remote:   data.Remote,
		DstPort:  data.DstPort,
		Secret:   data.Secret,
		Crypt:    data.Crypt,
	}
	if !w.cfg.AddOutput(output) {
		w.out.Info("WorkerImple.AddOutput %s already existed", output.Id())
		return
	}
	output.GenName()
	w.addOutput(w.cfg.Bridge.Name, output)
}

func (w *WorkerImpl) DelOutput(data schema.Output) {
	output, removed := w.cfg.DelOutput(&co.Output{Link: data.Device})
	if !removed {
		w.out.Info("WorkerImpl.DelOutput: %s not found", data.Device)
		return
	}
	w.delOutput(w.cfg.Bridge.Name, output)
}

func (w *WorkerImpl) SaveOutput() {
	w.cfg.SaveOutput()
}
func (w *WorkerImpl) ZTruster() api.ZTruster {
	return w.ztrust
}

func (w *WorkerImpl) Qoser() api.Qoser {
	return w.qos
}

func (w *WorkerImpl) ACLer() api.ACLer {
	return w.acl
}

func (w *WorkerImpl) FindHoper() api.FindHoper {
	return w.findhop
}
