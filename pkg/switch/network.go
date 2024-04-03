package cswitch

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
	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

func NewNetworker(c *co.Network) api.Networker {
	var obj api.Networker
	switch c.Provider {
	case "esp":
		obj = NewESPWorker(c)
	case "vxlan":
		obj = NewVxLANWorker(c)
	case "fabric":
		obj = NewFabricWorker(c)
	case "router":
		obj = NewRouterWorker(c)
	default:
		obj = NewOpenLANWorker(c)
	}
	api.AddWorker(c.Name, obj)
	return obj
}

type LinuxPort struct {
	cfg  co.Output
	link string
}

func (l *LinuxPort) String() string {
	return fmt.Sprintf("%s:%s:%d", l.cfg.Protocol, l.cfg.Remote, l.cfg.Segment)
}

type WorkerImpl struct {
	uuid    string
	cfg     *co.Network
	out     *libol.SubLogger
	dhcp    *Dhcp
	outputs []*LinuxPort
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
	rte := models.NewRoute(rt.Prefix, w.IfAddr(), rt.Mode)
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

	n := models.Network{
		Name:    cfg.Name,
		IpStart: cfg.Subnet.Start,
		IpEnd:   cfg.Subnet.End,
		Netmask: cfg.Subnet.Netmask,
		IfAddr:  cfg.Bridge.Address,
		Routes:  make([]*models.Route, 0, 2),
	}
	for _, rt := range cfg.Routes {
		nRoute := w.newRoute(&rt)
		if nRoute != nil {
			n.Routes = append(n.Routes, nRoute)
		}
	}

	cache.Network.Add(&n)

	w.updateVPN()
	w.createVPN()

	w.fire = cn.NewFireWallTable(cfg.Name)

	if out, err := w.setV.Clear(); err != nil {
		w.out.Error("WorkImpl.Initialize: create ipset: %s %s", out, err)
	}
	if out, err := w.setR.Clear(); err != nil {
		w.out.Error("WorkImpl.Initialize: create ipset: %s %s", out, err)
	}

	if cfg.ZTrust == "enable" {
		w.ztrust = NewZTrust(cfg.Name, 30)
		w.ztrust.Initialize()
	}

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

func (w *WorkerImpl) addOutput(bridge string, port *LinuxPort) {
	cfg := port.cfg
	out := &models.Output{
		Network:  w.cfg.Name,
		NewTime:  time.Now().Unix(),
		Protocol: cfg.Protocol,
		Remote:   cfg.Remote,
		Segment:  cfg.Segment,
	}

	mtu := 0
	if cfg.Protocol == "gre" {
		if port.link == "" {
			port.link = co.GenName("gre")
		}
		mtu = 1450
		link := &nl.Gretap{
			IKey: uint32(cfg.Segment),
			OKey: uint32(cfg.Segment),
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
				MTU:  mtu,
			},
			Local:    libol.ParseAddr("0.0.0.0"),
			Remote:   libol.ParseAddr(cfg.Remote),
			PMtuDisc: 1,
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", port.String(), err)
			return
		}
	} else if cfg.Protocol == "vxlan" {
		if port.link == "" {
			port.link = co.GenName("vxn")
		}
		dport := 8472
		if cfg.DstPort > 0 {
			dport = cfg.DstPort
		}
		mtu = 1450
		link := &nl.Vxlan{
			VxlanId: cfg.Segment,
			LinkAttrs: nl.LinkAttrs{
				TxQLen: -1,
				Name:   port.link,
				MTU:    mtu,
			},
			Group: libol.ParseAddr(cfg.Remote),
			Port:  dport,
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", port.String(), err)
			return
		}
	} else {
		link, err := nl.LinkByName(cfg.Remote)
		if link == nil {
			w.out.Error("WorkerImpl.addOutput %s %s", cfg.Remote, err)
			return
		}
		if err := nl.LinkSetUp(link); err != nil {
			w.out.Warn("WorkerImpl.addOutput %s %s", cfg.Remote, err)
		}

		if cfg.Segment > 0 {
			if port.link == "" {
				port.link = fmt.Sprintf("%s.%d", cfg.Remote, cfg.Segment)
			}
			subLink := &nl.Vlan{
				LinkAttrs: nl.LinkAttrs{
					Name:        port.link,
					ParentIndex: link.Attrs().Index,
				},
				VlanId: cfg.Segment,
			}
			if err := nl.LinkAdd(subLink); err != nil {
				w.out.Error("WorkerImpl.linkAdd %s %s", subLink.Name, err)
				return
			}
		} else {
			port.link = cfg.Remote
		}
	}

	if mtu > 0 {
		if w.br != nil {
			w.br.SetMtu(mtu)
		}
	}

	out.Device = port.link
	cache.Output.Add(port.link, out)

	w.out.Info("WorkerImpl.addOutput %s %s", port.link, port.String())
	w.AddPhysical(bridge, port.link)
}

func (w *WorkerImpl) loadRoute(rt co.PrefixRoute) {
	// install routes
	ifAddr := w.IfAddr()

	dst, err := libol.ParseNet(rt.Prefix)
	if err != nil {
		return
	}
	if ifAddr == rt.NextHop && rt.MultiPath == nil {
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
	w.out.Debug("WorkerImpl.loadRoute: %s", nlr.String())
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
	w.out.Debug("WorkerImpl.LoadRoute: %v", cfg.Routes)

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

func (w *WorkerImpl) Start(v api.Switcher) {
	cfg, vpn := w.GetCfgs()
	fire := w.fire

	w.out.Info("WorkerImpl.Start")

	w.loadVRF()
	w.loadRoutes()

	w.acl.Start()
	w.toACL(cfg.Bridge.Name)

	if cfg.Bridge.Mss > 0 {
		// forward to remote
		fire.Mangle.Post.AddRule(cn.IPRule{
			Output:  cfg.Bridge.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    cn.CTcpMss,
			SetMss:  cfg.Bridge.Mss,
		})
		// connect from local
		fire.Mangle.In.AddRule(cn.IPRule{
			Input:   cfg.Bridge.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    cn.CTcpMss,
			SetMss:  cfg.Bridge.Mss,
		})
	}

	for _, output := range cfg.Outputs {
		port := &LinuxPort{
			cfg: output,
		}
		w.addOutput(cfg.Bridge.Name, port)
		w.outputs = append(w.outputs, port)
	}

	if !(w.dhcp == nil) {
		w.dhcp.Start()
	}

	if !(w.vpn == nil) {
		w.vpn.Start()
		if !(w.vrf == nil) {
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

		if !(w.ztrust == nil) {
			w.ztrust.Start()
			fire.Mangle.Pre.AddRule(cn.IPRule{
				Input:   vpn.Device,
				CtState: "RELATED,ESTABLISHED",
				Comment: "Forwarding Accpted",
			})
			fire.Mangle.Pre.AddRule(cn.IPRule{
				Input:   vpn.Device,
				Jump:    w.ztrust.Chain(),
				Comment: "Goto Zero Trust",
			})
		}

		if !(w.qos == nil) {
			w.qos.Start()

			fire.Mangle.In.AddRule(cn.IPRule{
				Input:   vpn.Device,
				Jump:    w.qos.ChainIn(),
				Comment: "Goto Qos ChainIn",
			})
		}
	}

	fire.Start()
}

func (w *WorkerImpl) DelPhysical(bridge string, output string) {
	br := cn.NewBrCtl(bridge, 0)
	if err := br.DelPort(output); err != nil {
		w.out.Warn("WorkerImpl.DelPhysical %s", err)
	}
}

func (w *WorkerImpl) delOutput(bridge string, port *LinuxPort) {
	cfg := port.cfg
	w.out.Info("WorkerImpl.delOutput %s %s", port.link, port.String())

	cache.Output.Del(port.link)
	w.DelPhysical(bridge, port.link)

	if cfg.Protocol == "gre" {
		link := &nl.Gretap{
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
			},
		}
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.LinkDel %s %s", link.Name, err)
			return
		}
	} else if cfg.Protocol == "vxlan" {
		link := &nl.Vxlan{
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
			},
		}
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.LinkDel %s %s", link.Name, err)
			return
		}
	} else if port.cfg.Segment > 0 {
		link := &nl.Vlan{
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
			},
		}

		if err := nl.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.LinkDel %s %s", link.Name, err)
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

func (w *WorkerImpl) RestartVpn() {
	if w.vpn != nil {
		w.vpn.Restart()
	}
}

func (w *WorkerImpl) Stop() {
	w.out.Info("WorkerImpl.Stop")

	w.fire.Stop()
	w.unloadRoutes()

	if !(w.vpn == nil) {
		if !(w.ztrust == nil) {
			w.ztrust.Stop()
		}
		if !(w.qos == nil) {
			w.qos.Stop()
		}

		w.vpn.Stop()

	}

	if !(w.dhcp == nil) {
		w.dhcp.Stop()
	}

	for _, output := range w.outputs {
		w.delOutput(w.cfg.Bridge.Name, output)
	}
	w.outputs = nil

	w.acl.Stop()

	w.setR.Destroy()
	w.setV.Destroy()

}

func (w *WorkerImpl) String() string {
	return w.cfg.Name
}

func (w *WorkerImpl) ID() string {
	return w.uuid
}

func (w *WorkerImpl) Bridge() cn.Bridger {
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
	w.fire.Nat.Post.AddRule(cn.IPRule{
		Mark:    uint32(w.table),
		Source:  source,
		DestSet: pfxSet,
		Output:  output,
		Jump:    cn.CMasq,
		Comment: comment,
	})

}

func (w *WorkerImpl) toMasq_s(srcSet, prefix, comment string) {
	output := ""
	// Enable masquerade from source to prefix.
	w.fire.Nat.Post.AddRule(cn.IPRule{
		Mark:    uint32(w.table),
		SrcSet:  srcSet,
		Dest:    prefix,
		Output:  output,
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
		vpn.Push = append(vpn.Push, "redirect-gateway def1")
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

func (w *WorkerImpl) forwardVPNIpSet(rt string) {
	if rt == "0.0.0.0/0" {
		w.setV.Add("0.0.0.0/1")
		w.setV.Add("128.0.0.0/1")
		return
	}
	w.setV.Add(rt)
}

func (w *WorkerImpl) delForwardVPNIpSet(rt string) {
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
		w.forwardVPNIpSet(rt)
	}
	if w.vrf != nil {
		w.toForward_r(w.vrf.Name(), vpn.Subnet, w.setV.Name, "From VPN")
	} else {
		w.toForward_r(devName, vpn.Subnet, w.setV.Name, "From VPN")
	}
	w.toMasq_r(vpn.Subnet, w.setV.Name, "From VPN")
}

func (w *WorkerImpl) forwardSubnetIpSet(rt co.PrefixRoute) bool {
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

func (w *WorkerImpl) delForwardIpSet(rt co.PrefixRoute) {
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

	subnet := w.Subnet()
	// Enable MASQUERADE, and FORWARD it.
	w.toRelated(input, "Accept related")
	for _, rt := range cfg.Routes {
		if !w.forwardSubnetIpSet(rt) {
			break
		}
	}

	if w.vrf != nil {
		w.toForward_r(w.vrf.Name(), subnet.String(), w.setR.Name, "To route")
	} else {
		w.toForward_r(input, subnet.String(), w.setR.Name, "To route")
	}

	if vpn != nil {
		w.toMasq_s(w.setR.Name, vpn.Subnet, "To VPN")
	}

	w.toMasq_r(subnet.String(), w.setR.Name, "To Masq")
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

func (w *WorkerImpl) delCacheRoute(rt co.PrefixRoute) {
	if rt.NextHop == "" {
		w.out.Warn("WorkerImpl.DelCacheRoute: %s noNextHop", rt.Prefix)
		return
	}
	rte := models.NewRoute(rt.Prefix, w.IfAddr(), rt.Mode)
	if rt.Metric > 0 {
		rte.Metric = rt.Metric
	}
	if rt.NextHop != "" {
		rte.Origin = rt.NextHop
	}

	cache.Network.DelRoute(w.cfg.Name, rt)
}

func (w *WorkerImpl) addCacheRoute(rt co.PrefixRoute) {
	if rt.NextHop == "" {
		w.out.Warn("WorkerImpl.AddCacheRoute: %s ", rt.Prefix)
		return
	}
	rte := models.NewRoute(rt.Prefix, w.IfAddr(), rt.Mode)
	if rt.Metric > 0 {
		rte.Metric = rt.Metric
	}
	if rt.NextHop != "" {
		rte.Origin = rt.NextHop
	}
	cache.Network.AddRoute(w.cfg.Name, rte)
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
		for i, s := range vpn.Push {
			if s == "redirect-gateway def1" {
				vpn.Push = append(vpn.Push[:i], vpn.Push[i+1:]...)
				break
			}
		}
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
		Mode:    route.Mode,
		Metric:  route.Metric,
	}

	br := w.cfg.Bridge
	ipAddr := ""
	if _i, _, err := net.ParseCIDR(br.Address); err == nil {
		ipAddr = _i.String()
	}

	rt.CorrectRoute(ipAddr)

	return rt
}

func (w *WorkerImpl) findRoute(rt co.PrefixRoute) (co.PrefixRoute, int) {
	for i, ert := range w.cfg.Routes {
		if ert.Prefix == rt.Prefix && rt.NextHop == ert.NextHop {
			return ert, i
		}
	}
	return co.PrefixRoute{}, -1
}

func (w *WorkerImpl) AddRoute(route *schema.PrefixRoute, switcher api.Switcher) error {

	rt := w.correctRoute(route)

	if _, index := w.findRoute(rt); index != -1 {
		w.out.Warn("WorkerImpl.AddRoute: route exist")
		return nil
	}

	w.cfg.Routes = append(w.cfg.Routes, rt)

	libol.Info("WorkerImpl.AddRoute: %v", rt)

	w.forwardSubnetIpSet(rt)

	if inet, err := libol.ParseNet(rt.Prefix); err == nil {
		w.forwardVPNIpSet(inet.String())
	}

	w.addCacheRoute(rt)
	w.addVPNRoute(rt)
	w.loadRoute(rt)

	return nil
}

func (w *WorkerImpl) DelRoute(route *schema.PrefixRoute, switcher api.Switcher) error {

	correctRt := w.correctRoute(route)

	delRt, index := w.findRoute(correctRt)

	if index == -1 {

		w.out.Warn("WorkerImpl.DelRoute: route not found")
		return nil
	}

	w.cfg.Routes = append(w.cfg.Routes[:index], w.cfg.Routes[index+1:]...)

	w.delForwardIpSet(delRt)

	if inet, err := libol.ParseNet(delRt.Prefix); err == nil {
		w.delForwardVPNIpSet(inet.String())
	}

	w.delCacheRoute(delRt)
	w.delVPNRoute(delRt)
	w.unloadRoute(delRt)

	return nil
}

func (w *WorkerImpl) SaveRoute() {
	w.cfg.SaveRoute()
}

func (w *WorkerImpl) ZTruster() api.ZTruster {
	return w.ztrust
}

func (w *WorkerImpl) Qoser() api.Qoser {
	return w.qos
}

func (w *WorkerImpl) Router() api.Router {
	return w
}

func (w *WorkerImpl) IfAddr() string {
	return strings.SplitN(w.cfg.Bridge.Address, "/", 2)[0]
}

func (w *WorkerImpl) ACLer() api.ACLer {
	return w.acl
}

func (w *WorkerImpl) AddOutput(data schema.Output) {
	output := co.Output{
		Segment:  data.Segment,
		Protocol: data.Protocol,
		Remote:   data.Remote,
		DstPort:  data.DstPort,
	}
	w.cfg.Outputs = append(w.cfg.Outputs, output)
	port := &LinuxPort{
		cfg: output,
	}
	w.addOutput(w.cfg.Bridge.Name, port)
	w.outputs = append(w.outputs, port)
}

func (w *WorkerImpl) DelOutput(device string) {
	var linuxport *LinuxPort
	for _, v := range w.outputs {
		if v.link == device {
			linuxport = v
			break
		}
	}
	if linuxport == nil {
		return
	}
	Outputs := make([]co.Output, 0, len(w.cfg.Outputs))
	for _, v := range w.cfg.Outputs {
		if v != linuxport.cfg {
			Outputs = append(Outputs, v)
		}
	}
	w.cfg.Outputs = Outputs
	w.delOutput(w.cfg.Bridge.Name, linuxport)
	outputs := make([]*LinuxPort, 0, len(w.outputs))
	for _, v := range w.outputs {
		if v != linuxport {
			outputs = append(outputs, v)
		}
	}
	w.outputs = outputs
}

func (w *WorkerImpl) SaveOutput() {
	w.cfg.SaveOutput()
}
