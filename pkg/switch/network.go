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
	name string // gre:xx, vxlan:xx
	vlan int
	link string
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
	vrf     *cn.VRF
	table   int
	br      cn.Bridger
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

func (w *WorkerImpl) Initialize() {
	cfg := w.cfg

	if cfg.Namespace != "" {
		w.vrf = cn.NewVRF(cfg.Namespace, 0)
		w.table = w.vrf.Table()
	}

	n := models.Network{
		Name:    cfg.Name,
		IpStart: cfg.Subnet.Start,
		IpEnd:   cfg.Subnet.End,
		Netmask: cfg.Subnet.Netmask,
		IfAddr:  cfg.Bridge.Address,
		Routes:  make([]*models.Route, 0, 2),
	}
	for _, rt := range cfg.Routes {
		if rt.NextHop == "" {
			w.out.Warn("WorkerImpl.Initialize: %s noNextHop", rt.Prefix)
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

	w.updateVPN()
	w.createVPN()

	if cfg.Dhcp == "enable" {
		w.dhcp = NewDhcp(&co.Dhcp{
			Name:   cfg.Name,
			Subnet: cfg.Subnet,
			Bridge: cfg.Bridge,
		})
	}

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

	w.forwardSubnet()
	w.forwardVPN()
}

func (w *WorkerImpl) AddPhysical(bridge string, vlan int, output string) {
	link, err := nl.LinkByName(output)
	if err != nil {
		w.out.Error("WorkerImpl.LinkByName %s %s", output, err)
		return
	}
	slaver := output
	if vlan > 0 {
		if err := nl.LinkSetUp(link); err != nil {
			w.out.Warn("WorkerImpl.LinkSetUp %s %s", output, err)
		}
		subLink := &nl.Vlan{
			LinkAttrs: nl.LinkAttrs{
				Name:        fmt.Sprintf("%s.%d", output, vlan),
				ParentIndex: link.Attrs().Index,
			},
			VlanId: vlan,
		}
		if err := nl.LinkAdd(subLink); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", subLink.Name, err)
			return
		}
		slaver = subLink.Name
	}
	br := cn.NewBrCtl(bridge, 0)
	if err := br.AddPort(slaver); err != nil {
		w.out.Warn("WorkerImpl.AddPhysical %s", err)
	}
}

func (w *WorkerImpl) AddOutput(bridge string, port *LinuxPort) {
	name := port.name
	values := strings.SplitN(name, ":", 6)

	out := &models.Output{
		Network: w.cfg.Name,
		NewTime: time.Now().Unix(),
	}
	if values[0] == "gre" {
		if len(values) < 3 {
			w.out.Error("WorkerImpl.LinkAdd %s wrong", name)
			return
		}

		if port.link == "" {
			port.link = co.GenName("gre")
		}
		key, _ := strconv.Atoi(values[2])
		link := &nl.Gretap{
			IKey: uint32(key),
			OKey: uint32(key),
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
				MTU:  1460,
			},
			Local:    libol.ParseAddr("0.0.0.0"),
			Remote:   libol.ParseAddr(values[1]),
			PMtuDisc: 1,
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", name, err)
			return
		}
		out.Protocol = "gre"
		out.Connection = fmt.Sprintf("%s:%s", values[1], values[2])
	} else if values[0] == "vxlan" {
		if len(values) < 3 {
			w.out.Error("WorkerImpl.LinkAdd %s wrong", name)
			return
		}

		if port.link == "" {
			port.link = co.GenName("vxn")
		}
		dport := 8472
		if len(values) == 4 {
			dport, _ = strconv.Atoi(values[3])
		}
		vni, _ := strconv.Atoi(values[2])
		link := &nl.Vxlan{
			VxlanId: vni,
			LinkAttrs: nl.LinkAttrs{
				TxQLen: -1,
				Name:   port.link,
				MTU:    1450,
			},
			Group: libol.ParseAddr(values[1]),
			Port:  dport,
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", name, err)
			return
		}
		out.Protocol = "vxlan"
		out.Connection = fmt.Sprintf("%s:%s", values[1], values[2])
	} else {
		port.link = name
	}

	out.Device = port.link
	cache.Output.Add(port.link, out)

	w.out.Info("WorkerImpl.AddOutput %s %s", port.link, port.name)
	w.AddPhysical(bridge, port.vlan, port.link)
}

func (w *WorkerImpl) loadRoutes() {
	// install routes
	cfg := w.cfg
	w.out.Debug("WorkerImpl.LoadRoute: %v", cfg.Routes)
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
		nlrt := nl.Route{
			Dst:   dst,
			Table: w.table,
		}
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
		w.out.Debug("WorkerImpl.LoadRoute: %s", nlrt.String())
		promise := libol.NewPromise()
		rt_c := rt
		promise.Go(func() error {
			if err := nl.RouteReplace(&nlrt); err != nil {
				w.out.Warn("WorkerImpl.LoadRoute: %v %s", nlrt, err)
				return err
			}
			w.out.Info("WorkerImpl.LoadRoute: %v success", rt_c.String())
			return nil
		})
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

	if cfg.Acl != "" {
		fire.Mangle.Pre.AddRule(cn.IPRule{
			Input: cfg.Bridge.Name,
			Jump:  cfg.Acl,
		})
	}

	fire.Filter.For.AddRule(cn.IPRule{
		Input:  cfg.Bridge.Name,
		Output: cfg.Bridge.Name,
	})

	if cfg.Bridge.Mss > 0 {
		// forward to remote
		fire.Mangle.Post.AddRule(cn.IPRule{
			Output:  cfg.Bridge.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    "TCPMSS",
			SetMss:  cfg.Bridge.Mss,
		})
		// connect from local
		fire.Mangle.In.AddRule(cn.IPRule{
			Input:   cfg.Bridge.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    "TCPMSS",
			SetMss:  cfg.Bridge.Mss,
		})
	}

	for _, output := range cfg.Outputs {
		port := &LinuxPort{
			name: output.Interface,
			vlan: output.Vlan,
		}
		w.AddOutput(cfg.Bridge.Name, port)
		w.outputs = append(w.outputs, port)
	}

	if !(w.dhcp == nil) {
		w.dhcp.Start()
		fire.Nat.Post.AddRule(cn.IPRule{
			Source:  cfg.Bridge.Address,
			NoDest:  cfg.Bridge.Address,
			Jump:    cn.CMasq,
			Comment: "Default Gateway for DHCP",
		})
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

				_, dest, _ := net.ParseCIDR(vpn.Subnet)
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
	}

	if !(w.vpn == nil || w.ztrust == nil) {
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

	fire.Start()
}

func (w *WorkerImpl) DelPhysical(bridge string, vlan int, output string) {
	if vlan > 0 {
		subLink := &nl.Vlan{
			LinkAttrs: nl.LinkAttrs{
				Name: fmt.Sprintf("%s.%d", output, vlan),
			},
		}
		if err := nl.LinkDel(subLink); err != nil {
			w.out.Error("WorkerImpl.DelPhysical.LinkDel %s %s", subLink.Name, err)
			return
		}
	} else {
		br := cn.NewBrCtl(bridge, 0)
		if err := br.DelPort(output); err != nil {
			w.out.Warn("WorkerImpl.DelPhysical %s", err)
		}
	}
}

func (w *WorkerImpl) DelOutput(bridge string, port *LinuxPort) {
	w.out.Info("WorkerImpl.DelOutput %s %s", port.link, port.name)

	cache.Output.Del(port.link)
	w.DelPhysical(bridge, port.vlan, port.link)

	values := strings.SplitN(port.name, ":", 6)
	if values[0] == "gre" {
		link := &nl.Gretap{
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
			},
		}
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.DelOutput.LinkDel %s %s", link.Name, err)
			return
		}
	} else if values[0] == "vxlan" {
		link := &nl.Vxlan{
			LinkAttrs: nl.LinkAttrs{
				Name: port.link,
			},
		}
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.DelOutput.LinkDel %s %s", link.Name, err)
			return
		}
	}
}

func (w *WorkerImpl) unloadRoutes() {
	cfg := w.cfg
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nlRt := nl.Route{
			Dst:   dst,
			Table: w.table,
		}
		if rt.MultiPath == nil {
			nlRt.Gw = net.ParseIP(rt.NextHop)
			nlRt.Priority = rt.Metric
		}
		w.out.Debug("WorkerImpl.UnLoadRoute: %s", nlRt.String())
		if err := nl.RouteDel(&nlRt); err != nil {
			w.out.Warn("WorkerImpl.UnLoadRoute: %s", err)
			continue
		}
		w.out.Info("WorkerImpl.UnLoadRoute: %v", rt.String())
	}
}

func (w *WorkerImpl) Stop() {
	w.out.Info("WorkerImpl.Stop")

	w.fire.Stop()

	if !(w.vpn == nil || w.ztrust == nil) {
		w.ztrust.Stop()
	}

	if !(w.vpn == nil) {
		w.vpn.Stop()
	}

	if !(w.dhcp == nil) {
		w.dhcp.Stop()
	}

	for _, output := range w.outputs {
		w.DelOutput(w.cfg.Bridge.Name, output)
	}

	w.outputs = nil
	w.setR.Destroy()
	w.setV.Destroy()

	w.unloadRoutes()
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

func (w *WorkerImpl) Subnet() string {
	cfg := w.cfg

	ipAddr := cfg.Bridge.Address
	ipMask := cfg.Subnet.Netmask
	if ipAddr == "" {
		ipAddr = cfg.Subnet.Start
	}
	if ipAddr == "" {
		return ""
	}

	addr := ipAddr
	if ipMask != "" {
		prefix := libol.Netmask2Len(ipMask)
		ifAddr := strings.SplitN(ipAddr, "/", 2)[0]
		addr = fmt.Sprintf("%s/%d", ifAddr, prefix)
	}
	if _, inet, err := net.ParseCIDR(addr); err == nil {
		return inet.String()
	}

	return ""
}

func (w *WorkerImpl) Reload(v api.Switcher) {
}

func (w *WorkerImpl) toACL(acl, input string) {
	if input == "" {
		return
	}
	if acl != "" {
		w.fire.Mangle.Pre.AddRule(cn.IPRule{
			Input: input,
			Jump:  acl,
		})
	}
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

func (w *WorkerImpl) updateVPN() {
	cfg, vpn := w.GetCfgs()
	if vpn == nil {
		return
	}

	routes := vpn.Routes
	routes = append(routes, vpn.Subnet) // add subnet of VPN self.
	if addr := w.Subnet(); addr != "" {
		w.out.Info("WorkerImpl.updateVPN subnet %s", addr)
		routes = append(routes, addr)
	}

	for _, rt := range cfg.Routes {
		addr := rt.Prefix
		if addr == "0.0.0.0/0" {
			vpn.Push = append(vpn.Push, "redirect-gateway def1")
			routes = append(routes, addr)
			continue
		}
		if _, inet, err := net.ParseCIDR(addr); err == nil {
			routes = append(routes, inet.String())
		}
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
}

func (w *WorkerImpl) forwardVPN() {
	cfg, vpn := w.GetCfgs()
	if vpn == nil {
		return
	}

	devName := vpn.Device

	w.forwardZone(devName)

	_, port := libol.GetHostPort(vpn.Listen)
	if vpn.Protocol == "udp" {
		w.openPort("udp", port, "Open VPN")
	} else {
		w.openPort("tcp", port, "Open VPN")
	}

	// Enable MASQUERADE, and FORWARD it.
	w.toRelated(devName, "Accept related")
	w.toACL(cfg.Acl, devName)

	for _, rt := range vpn.Routes {
		if rt == "0.0.0.0/0" {
			w.setV.Add("0.0.0.0/1")
			w.setV.Add("128.0.0.0/1")
			break
		}
		w.setV.Add(rt)
	}

	w.toForward_r(devName, vpn.Subnet, w.setV.Name, "From VPN")
	w.toMasq_r(vpn.Subnet, w.setV.Name, "From VPN")
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
	w.toForward_r(input, subnet, w.setR.Name, "To route")
	if vpn != nil {
		w.toMasq_s(w.setR.Name, vpn.Subnet, "To VPN")
	}
	w.toMasq_r(subnet, w.setR.Name, "To Masq")
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

func (w *WorkerImpl) ZTruster() api.ZTruster {
	return w.ztrust
}

func (w *WorkerImpl) IfAddr() string {
	return strings.SplitN(w.cfg.Bridge.Address, "/", 2)[0]
}
