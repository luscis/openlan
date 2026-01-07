package cswitch

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

func NewNetworker(c *co.Network) api.NetworkApi {
	var obj api.NetworkApi

	switch c.Provider {
	case "ipsec":
		obj = NewIPSecWorker(c)
	case "bgp":
		obj = NewBgpWorker(c)
	case "router":
		obj = NewRouterWorker(c)
	case "ceci":
		obj = NewCeciWorker(c)
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
	ipser   *cn.IPSet
	vpn     *OpenVPN
	ztrust  *ZTrust
	qos     *QosCtrl
	vrf     *cn.VRF
	table   int
	br      cn.Bridger
	acl     *ACL
	findhop *FindHop
	snat    *cn.FireWallChain
	dnat    *cn.FireWallChain
}

func NewWorkerApi(c *co.Network) *WorkerImpl {
	return &WorkerImpl{
		cfg:   c,
		out:   libol.NewSubLogger(c.Name),
		ipser: cn.NewIPSet(c.Name+"_r", "hash:net"),
		table: 0,
	}
}

func (w *WorkerImpl) Provider() string {
	return w.cfg.Provider
}

func (w *WorkerImpl) addCache() {
	cfg := w.cfg
	n := models.Network{
		Name:   cfg.Name,
		Config: cfg,
	}
	if cfg.Subnet != nil {
		n.IpStart = cfg.Subnet.Start
		n.IpEnd = cfg.Subnet.End
		n.Netmask = cfg.Subnet.Netmask
		if cfg.Bridge != nil {
			n.Address = cfg.Bridge.Address
		}
	}
	cache.Network.Add(&n)
}

func (w *WorkerImpl) Initialize() {
	cfg := w.cfg

	if cfg.Namespace != "" {
		w.vrf = cn.NewVRF(cfg.Namespace, 0)
		w.table = w.vrf.Table()
	}

	w.acl = NewACL(cfg.Name)
	w.acl.Initialize()

	w.addCache()
	w.findhop = NewFindHop(cfg.Name, cfg)

	w.setVPN()
	w.newVPN()

	w.fire = cn.NewFireWallTable(cfg.Name)
	w.snat = cn.NewFireWallChain("XTT_"+cfg.Name+"_SNAT", cn.TNat, "")
	w.dnat = cn.NewFireWallChain("XTT_"+cfg.Name+"_DNAT", cn.TNat, "")

	w.ipser.Clear()

	w.ztrust = NewZTrust(cfg.Name, 30)
	w.ztrust.Initialize()

	w.qos = NewQosCtrl(cfg.Name)
	w.qos.Initialize()

	if cfg.Dhcp == "enable" && cfg.Bridge != nil {
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

	w.toSubnet()
	w.toVPN()
}

func (w *WorkerImpl) toSnat() {
	w.snat.Prepare()
	w.fire.Nat.Post.AddRuleX(cn.IPRule{
		Jump:    w.snat.Chain().Name,
		Comment: "Goto SNAT",
	})
}

func (w *WorkerImpl) doSnat() {
	w.snat.Flush()
	cfg, vpn := w.GetCfgs()
	if cfg.Snat == "disable" {
		return
	}
	if cfg.Snat == "enable" {
		w.toMasq_i("", w.ipser.Name, "To Masq")
	}
	if vpn != nil && (cfg.Snat == "openvpn" || cfg.Snat == "enable") {
		w.toMasq_r(vpn.Subnet, w.ipser.Name, "From VPN")
	}
}

func (w *WorkerImpl) AddPhysical(bridge string, output string) {
	br := cn.NewBrCtl(bridge, 0)
	if err := br.AddPort(output); err != nil {
		w.out.Warn("WorkerImpl.AddPhysical %s", err)
	}
}

func (w *WorkerImpl) addOutput(bridge string, port *co.Output) {
	mtu := 0
	out := &models.Output{
		Network: w.cfg.Name,
		NewTime: time.Now().Unix(),
	}

	switch port.Protocol {
	case "gre":
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
		link.Stop()
		if err := link.Start(); err != nil {
			w.out.Error("WorkerImpl.LinkStart %s %s", port.Id(), err)
			return
		}
		port.Linker = link
	case "vxlan":
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
		link.Stop()
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
	case "tcp", "tls", "wss":
		port.Link = cn.Taps.GenName()
		name, pass := SplitCombined(port.Secret)
		algo, secret := SplitCombined(port.Crypt)
		ac := co.Access{
			Alias:       w.cfg.Alias,
			Network:     w.cfg.Name,
			RequestAddr: false,
			Interface: co.Interface{
				Name:   port.Link,
				Bridge: bridge,
			},
			Connection: port.Remote,
			Fallback:   port.Fallback,
			Protocol:   port.Protocol,
			Username:   name,
			Password:   pass,
		}
		if port.DstPort != 0 {
			ac.Connection = fmt.Sprintf("%s:%d", port.Remote, port.DstPort)
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
		out.StatsFile = link.StatusFile()
	default:
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

	out.Protocol = port.Protocol
	out.Remote = port.Remote
	out.Segment = port.Segment
	out.Device = port.Link
	out.Secret = port.Secret
	out.Fallback = port.Fallback
	cache.Output.Add(port.Link, out)

	w.out.Info("WorkerImpl.addOutput %s %s", port.Link, port.Id())
	w.AddPhysical(bridge, port.Link)
}

func (w *WorkerImpl) toRoute(rt co.PrefixRoute) {
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
	w.out.Info("WorkerImpl.toRoute: %s", nlr.String())

	rt_c := rt
	promise := libol.NewPromise()
	promise.Go(func() error {
		if err := nl.RouteReplace(&nlr); err != nil {
			w.out.Warn("WorkerImpl.toRoute: %v %s", nlr, err)
			return err
		}
		w.out.Info("WorkerImpl.toRoute: %v success", rt_c.String())
		return nil
	})
}

func (w *WorkerImpl) toRoutes() {
	// install routes
	cfg := w.cfg
	w.out.Info("WorkerImpl.toRoute: %v", cfg.Routes)

	for _, rt := range cfg.Routes {
		w.toRoute(rt)
	}
}

func (w *WorkerImpl) toVRF() {
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

func (w *WorkerImpl) doMss() {
	cfg, _ := w.GetCfgs()

	if cfg.Bridge == nil || cfg.Bridge.Mss <= 0 {
		return
	}

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
	if cfg.Bridge == nil {
		return
	}

	if cfg.Bridge.Mss != mss {
		cfg.Bridge.Mss = mss
		w.doMss()
	}
}

func (w *WorkerImpl) SetSnat(value string) {
	cfg, _ := w.GetCfgs()
	switch value {
	case "enable", "disable", "openvpn":
		cfg.Snat = value
		w.doSnat()
	}
}

func (w *WorkerImpl) doDnat() {
	cfg, _ := w.GetCfgs()
	w.out.Info("WorkerImpl: doDnat")

	w.fire.Nat.Pre.AddRuleX(cn.IPRule{
		Jump:    w.dnat.Chain().Name,
		Comment: "Goto DNAT",
	})
	for _, obj := range cfg.Dnat {
		if err := w.dnat.AddRuleX(cn.IPRule{
			Proto:   obj.Protocol,
			Dest:    obj.Dest,
			DstPort: fmt.Sprintf("%d", obj.Dport),
			ToDest:  fmt.Sprintf("%s:%d", obj.ToDest, obj.ToDport),
			Jump:    "DNAT",
			Comment: "DNAT " + obj.Id(),
		}); err != nil {
			w.out.Warn("WorkerImple: doDnat: %s", err)
		}
	}
}

func (w *WorkerImpl) AddDnat(data schema.DNAT) error {
	cfg, _ := w.GetCfgs()
	obj := config.Dnat{
		Protocol: data.Protocol,
		Dest:     data.Dest,
		Dport:    data.Dport,
		ToDest:   data.ToDest,
		ToDport:  data.ToDport,
	}
	obj.Correct()

	if ok := cfg.AddDnat(&obj); ok {
		if err := w.dnat.AddRuleX(cn.IPRule{
			Proto:   obj.Protocol,
			Dest:    obj.Dest,
			DstPort: fmt.Sprintf("%d", obj.Dport),
			ToDest:  fmt.Sprintf("%s:%d", obj.ToDest, obj.ToDport),
			Jump:    "DNAT",
			Comment: "DNAT " + obj.Id(),
		}); err != nil {
			w.out.Warn("WorkerImple: AddDnat: %s", err)
		}
	}
	return nil
}

func (w *WorkerImpl) DelDnat(data schema.DNAT) error {
	cfg, _ := w.GetCfgs()
	obj := config.Dnat{
		Protocol: data.Protocol,
		Dest:     data.Dest,
		Dport:    data.Dport,
		ToDest:   data.ToDest,
		ToDport:  data.ToDport,
	}
	obj.Correct()

	if older, ok := cfg.DelDnat(&obj); ok {
		if err := w.dnat.DelRuleX(cn.IPRule{
			Proto:   older.Protocol,
			Dest:    older.Dest,
			DstPort: fmt.Sprintf("%d", older.Dport),
			ToDest:  fmt.Sprintf("%s:%d", older.ToDest, older.ToDport),
			Jump:    "DNAT",
			Comment: "DNAT " + older.Id(),
		}); err != nil {
			w.out.Warn("WorkerImple: DelDnat: %s", err)
		}
	}
	return nil
}

func (w *WorkerImpl) ListDnat(call func(data schema.DNAT)) {
	cfg, _ := w.GetCfgs()
	cfg.ListDnat(func(value config.Dnat) {
		call(schema.DNAT{
			Protocol: value.Protocol,
			Dest:     value.Dest,
			Dport:    value.Dport,
			ToDest:   value.ToDest,
			ToDport:  value.ToDport,
		})
	})
}

func (w *WorkerImpl) toTrust() {
	_, vpn := w.GetCfgs()
	w.fire.Mangle.Pre.AddRuleX(cn.IPRule{
		Input:   vpn.Device,
		Jump:    w.ztrust.Chain(),
		Comment: "Goto Zero Trust",
	})
}

func (w *WorkerImpl) leftTrust() {
	_, vpn := w.GetCfgs()
	w.fire.Mangle.Pre.DelRuleX(cn.IPRule{
		Input:   vpn.Device,
		Jump:    w.ztrust.Chain(),
		Comment: "Goto Zero Trust",
	})
}

func (w *WorkerImpl) EnableZTrust() {
	cfg, _ := w.GetCfgs()
	if cfg.ZTrust != "enable" {
		cfg.ZTrust = "enable"
		w.toTrust()
	}
}

func (w *WorkerImpl) DisableZTrust() {
	cfg, _ := w.GetCfgs()
	if cfg.ZTrust == "enable" {
		cfg.ZTrust = "disable"
		w.leftTrust()
	}
}

func (w *WorkerImpl) setVPN2VRF() {
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
		w.out.Debug("WorkerImpl.toRoute: %s", rt.String())
		if err := nl.RouteAdd(rt); err != nil {
			w.out.Warn("Route add: %s", err)
			return err
		}

		return nil
	})
}

func (w *WorkerImpl) toVPNQoS() {
	_, vpn := w.GetCfgs()
	w.fire.Mangle.In.AddRuleX(cn.IPRule{
		Input:   vpn.Device,
		Jump:    w.qos.ChainIn(),
		Comment: "Goto Qos ChainIn",
	})
}

func (w *WorkerImpl) leftVPNQoS() {
	_, vpn := w.GetCfgs()
	w.fire.Mangle.In.DelRuleX(cn.IPRule{
		Input:   vpn.Device,
		Jump:    w.qos.ChainIn(),
		Comment: "Goto Qos ChainIn",
	})
}

func (w *WorkerImpl) Start(v api.SwitchApi) {
	cfg, _ := w.GetCfgs()

	w.out.Info("WorkerImpl.Start")

	w.toVRF()
	w.toRoutes()

	w.acl.Start()

	if cfg.Bridge != nil {
		w.toACL(cfg.Bridge.Name)
		for _, output := range cfg.Outputs {
			w.addOutput(cfg.Bridge.Name, output)
		}
	}

	if !(w.vpn == nil) {
		w.vpn.Start()
		if !(w.vrf == nil) {
			w.setVPN2VRF()
		}
		w.toVPNQoS()
		w.qos.Start()
		w.ztrust.Start()
	}

	w.fire.Start()
	w.snat.Install()
	w.dnat.Install()

	w.doDnat()
	w.toSnat()
	w.doSnat()

	if cfg.Bridge != nil {
		// forward to remote
		w.doMss()
	}

	w.findhop.Start()

	if !(w.dhcp == nil) {
		w.dhcp.Start()
	}
	if !(w.vpn == nil) {
		if cfg.ZTrust == "enable" {
			w.toTrust()
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

func (w *WorkerImpl) leftRoute(rt co.PrefixRoute) {
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
	w.out.Debug("WorkerImpl.leftRoute: %s", nlr.String())
	if err := nl.RouteDel(&nlr); err != nil {
		w.out.Warn("WorkerImpl.leftRoute: %s", err)
		return
	}
	w.out.Info("WorkerImpl.leftRoute: %v", rt.String())
}

func (w *WorkerImpl) leftRoutes() {
	cfg := w.cfg
	for _, rt := range cfg.Routes {
		w.leftRoute(rt)
	}
}

func (w *WorkerImpl) AddVPN(value schema.OpenVPN) error {
	if w.vpn != nil {
		return libol.NewErr("openvpn is running")
	}
	cfg := &co.OpenVPN{
		Listen:   value.Listen,
		Protocol: value.Protocol,
		Push:     value.Push,
		Subnet:   value.Subnet,
	}
	cfg.Correct(w.cfg.AddrPool, w.cfg.Name)
	w.cfg.OpenVPN = cfg

	w.setVPN()
	w.newVPN()
	w.toVPN()
	w.vpn.Start()
	if !(w.vrf == nil) {
		w.setVPN2VRF()
	}
	w.toVPNQoS()
	if w.cfg.ZTrust == "enable" {
		w.toTrust()
	}
	w.doSnat()
	return nil
}

func (w *WorkerImpl) DelVPN() {
	if w.vpn != nil {
		w.leftVPN()
		w.leftVPNQoS()
		w.vpn.Stop()
		w.vpn.CheckWait()
		if w.cfg.ZTrust == "enable" {
			w.leftTrust()
		}
		w.vpn = nil
		w.cfg.OpenVPN = nil
		w.doSnat()
	}
}

func (w *WorkerImpl) StartVPN() {
	if w.vpn == nil {
		return
	}

	w.vpn.Stop()
	w.vpn.CheckWait()
	w.vpn.Initialize()
	w.vpn.Start()
	if !(w.vrf == nil) {
		w.setVPN2VRF()
	}
}

func (w *WorkerImpl) AddVPNClient(name, address string) error {
	vpn := w.vpn
	if vpn == nil {
		return libol.NewErr("VPN was disabled")
	}

	return vpn.AddClient(name, address)
}

func (w *WorkerImpl) DelVPNClient(name string) error {
	vpn := w.vpn
	if vpn == nil {
		return libol.NewErr("VPN was disabled")
	}

	return vpn.DelClient(name)
}

func (w *WorkerImpl) ListClients(call func(name, local string)) {
	vpn := w.vpn
	if vpn == nil {
		return
	}

	vpn.ListClients(call)
}

func (w *WorkerImpl) Stop() {
	w.out.Info("WorkerImpl.Stop")

	cfg, _ := w.GetCfgs()
	w.snat.Cancel()
	w.dnat.Cancel()
	w.fire.Stop()
	w.findhop.Stop()
	w.acl.Stop()

	w.leftRoutes()

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

	if cfg.Bridge != nil {
		for _, output := range cfg.Outputs {
			w.delOutput(cfg.Bridge.Name, output)
		}
	}

	w.ipser.Destroy()
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

	if cfg.Bridge == nil {
		return nil
	}

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

func (w *WorkerImpl) Reload(v api.SwitchApi) {
}

func (w *WorkerImpl) toACL(input string) {
	if input != "" {
		w.fire.Raw.Pre.AddRuleX(cn.IPRule{
			Input: input,
			Jump:  w.acl.Chain(),
		})
	}

}

func (w *WorkerImpl) leftACL(input string) {
	if input != "" {
		w.fire.Raw.Pre.DelRuleX(cn.IPRule{
			Input: input,
			Jump:  w.acl.Chain(),
		})
	}
}

func (w *WorkerImpl) openPort(protocol, port, comment string) {
	w.out.Info("WorkerImpl.openPort %s %s", protocol, port)
	// allowed forward between source and prefix.
	w.fire.Filter.In.AddRuleX(cn.IPRule{
		Proto:   protocol,
		Match:   "multiport",
		DstPort: port,
		Comment: comment,
	})
}

func (w *WorkerImpl) closePort(protocol, port, comment string) {
	w.out.Info("WorkerImpl.ClosePort %s %s", protocol, port)
	// allowed forward between source and prefix.
	w.fire.Filter.In.DelRuleX(cn.IPRule{
		Proto:   protocol,
		Match:   "multiport",
		DstPort: port,
		Comment: comment,
	})
}

func (w *WorkerImpl) toForward_i(input, pfxSet, comment string) {
	w.out.Debug("WorkerImpl.toForward %s %s:%s", input, pfxSet)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.AddRuleX(cn.IPRule{
		Input:   input,
		DestSet: pfxSet,
		Comment: comment,
	})
}

func (w *WorkerImpl) toForward_r(input, source, pfxSet, comment string) {
	w.out.Debug("WorkerImpl.toForward %s:%s %s:%s", input, source, pfxSet)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.AddRuleX(cn.IPRule{
		Input:   input,
		Source:  source,
		DestSet: pfxSet,
		Comment: comment,
	})
}

func (w *WorkerImpl) leftForward_r(input, source, pfxSet, comment string) {
	w.out.Debug("WorkerImpl.leftForward %s:%s %s:%s", input, source, pfxSet)
	// Allowed forward between source and prefix.
	w.fire.Filter.For.DelRuleX(cn.IPRule{
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
	w.snat.AddRuleX(cn.IPRule{
		Mark:    uint32(w.table),
		Source:  source,
		DestSet: pfxSet,
		Output:  output,
		Jump:    cn.CMasq,
		Comment: comment,
	})
}

func (w *WorkerImpl) leftMasq_r(source, pfxSet, comment string) {
	// Enable masquerade from source to prefix.
	output := ""
	w.snat.DelRuleX(cn.IPRule{
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
	w.snat.AddRuleX(cn.IPRule{
		Mark:    uint32(w.table),
		Input:   input,
		DestSet: pfxSet,
		Jump:    cn.CMasq,
		Comment: comment,
	})

}

func (w *WorkerImpl) leftMasq_i(input, pfxSet, comment string) {
	// Enable masquerade from input to prefix.
	w.snat.DelRuleX(cn.IPRule{
		Mark:    uint32(w.table),
		Input:   input,
		DestSet: pfxSet,
		Jump:    cn.CMasq,
		Comment: comment,
	})
}

func (w *WorkerImpl) toMasq_s(srcSet, prefix, comment string) {
	// Enable masquerade from source to prefix.
	w.snat.AddRuleX(cn.IPRule{
		Mark:    uint32(w.table),
		SrcSet:  srcSet,
		Dest:    prefix,
		Jump:    cn.CMasq,
		Comment: comment,
	})
}

func (w *WorkerImpl) leftMasq_s(srcSet, prefix, comment string) {
	// Enable masquerade from source to prefix.
	w.snat.AddRuleX(cn.IPRule{
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
	if output != "" {
		w.fire.Filter.For.AddRuleX(cn.IPRule{
			Output:  output,
			CtState: "RELATED,ESTABLISHED",
			Comment: comment,
		})
	}
	w.fire.Filter.For.AddRuleX(cn.IPRule{
		Input:   output,
		CtState: "RELATED,ESTABLISHED",
		Comment: comment,
	})
}

func (w *WorkerImpl) leftRelated(output, comment string) {
	w.out.Debug("WorkerImpl.leftRelated %s", output)
	// Allowed forward between source and prefix.
	if output != "" {
		w.fire.Filter.For.DelRuleX(cn.IPRule{
			Output:  output,
			CtState: "RELATED,ESTABLISHED",
			Comment: comment,
		})
	}
	w.fire.Filter.For.DelRuleX(cn.IPRule{
		Input:   output,
		CtState: "RELATED,ESTABLISHED",
		Comment: comment,
	})
}

func (w *WorkerImpl) GetCfgs() (*co.Network, *co.OpenVPN) {
	cfg := w.cfg
	vpn := cfg.OpenVPN
	return cfg, vpn
}

func (w *WorkerImpl) setVPNRoute(routes []string, rt co.PrefixRoute) []string {
	_, vpn := w.GetCfgs()
	if vpn == nil {
		return routes
	}

	addr := rt.Prefix
	if addr == "0.0.0.0/0" {
		vpn.AddRedirectDef1()
		return routes
	}
	if inet, err := libol.ParseNet(addr); err == nil {
		routes = append(routes, inet.String())
	}

	return routes
}

func (w *WorkerImpl) setVPN() {
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
		routes = w.setVPNRoute(routes, rt)
	}
	vpn.Routes = routes
}

func (w *WorkerImpl) toZone(input string) {
	if w.table == 0 {
		return
	}

	w.out.Debug("WorkerImpl.toZone %s", input)
	w.fire.Raw.Pre.AddRuleX(cn.IPRule{
		Input:   input,
		Jump:    cn.CMark,
		SetMark: uint32(w.table),
		Comment: "Mark private traffic",
	})
	w.fire.Raw.Pre.AddRuleX(cn.IPRule{
		Input:   input,
		Jump:    cn.CCT,
		Zone:    uint32(w.table),
		Comment: "Goto private zone",
	})
	w.fire.Raw.Out.AddRuleX(cn.IPRule{
		Output:  input,
		Jump:    cn.CCT,
		Zone:    uint32(w.table),
		Comment: "Goto private zone",
	})
}

func (w *WorkerImpl) leftZone(input string) {
	if w.table == 0 {
		return
	}

	w.out.Debug("WorkerImpl.leftZone %s", input)
	w.fire.Raw.Pre.DelRuleX(cn.IPRule{
		Input:   input,
		Jump:    cn.CMark,
		SetMark: uint32(w.table),
		Comment: "Mark private traffic",
	})
	w.fire.Raw.Pre.DelRuleX(cn.IPRule{
		Input:   input,
		Jump:    cn.CCT,
		Zone:    uint32(w.table),
		Comment: "Goto private zone",
	})
	w.fire.Raw.Out.DelRuleX(cn.IPRule{
		Output:  input,
		Jump:    cn.CCT,
		Zone:    uint32(w.table),
		Comment: "Goto private zone",
	})
}

func (w *WorkerImpl) toVPN() {
	if _, vpn := w.GetCfgs(); vpn != nil {
		devName := vpn.Device
		_, port := libol.GetHostPort(vpn.Listen)
		if vpn.Protocol == "udp" {
			w.openPort("udp", port, "Open VPN")
		} else {
			w.openPort("tcp", port, "Open VPN")
		}

		w.toZone(devName)
		// Enable MASQUERADE, and FORWARD it.
		w.toRelated(devName, "Accept related")
		w.toACL(devName)
		w.addIPSet(co.PrefixRoute{Prefix: vpn.Subnet})
		if w.vrf != nil {
			w.toForward_r(w.vrf.Name(), vpn.Subnet, w.ipser.Name, "From VPN")
		} else {
			w.toForward_r(devName, vpn.Subnet, w.ipser.Name, "From VPN")
		}
	}
}

func (w *WorkerImpl) leftVPN() {
	if _, vpn := w.GetCfgs(); vpn != nil {
		devName := vpn.Device
		_, port := libol.GetHostPort(vpn.Listen)
		if vpn.Protocol == "udp" {
			w.closePort("udp", port, "Open VPN")
		} else {
			w.closePort("tcp", port, "Open VPN")
		}

		w.leftZone(devName)
		// disable MASQUERADE, and FORWARD.
		w.leftRelated(devName, "Accept related")
		w.leftACL(devName)
		w.delIPSet(co.PrefixRoute{Prefix: vpn.Subnet})
		if w.vrf != nil {
			w.leftForward_r(w.vrf.Name(), vpn.Subnet, w.ipser.Name, "From VPN")
		} else {
			w.leftForward_r(devName, vpn.Subnet, w.ipser.Name, "From VPN")
		}
	}
}

func (w *WorkerImpl) addIPSet(rt co.PrefixRoute) {
	if rt.MultiPath != nil {
		return
	}

	if rt.Prefix == "0.0.0.0/0" {
		w.ipser.Add("0.0.0.0/1")
		w.ipser.Add("128.0.0.0/1")
	}
	w.ipser.Add(rt.Prefix)
}

func (w *WorkerImpl) delIPSet(rt co.PrefixRoute) {
	if rt.MultiPath != nil {
		return
	}

	if rt.Prefix == "0.0.0.0/0" {
		w.ipser.Del("0.0.0.0/1")
		w.ipser.Del("128.0.0.0/1")
		return
	}
	w.ipser.Del(rt.Prefix)
}

func (w *WorkerImpl) toSubnet() {
	cfg, _ := w.GetCfgs()
	if cfg.Bridge != nil {
		input := cfg.Bridge.Name
		if w.br != nil {
			input = w.br.L3Name()
			w.toZone(input)
		}
		ifAddr := strings.SplitN(cfg.Bridge.Address, "/", 2)[0]
		if ifAddr == "" {
			return
		}
		// Enable MASQUERADE, and FORWARD it.
		w.toRelated(input, "Accept related")
	}
	for _, rt := range cfg.Routes {
		w.addIPSet(rt)
	}

	if w.vrf != nil {
		w.toForward_i(w.vrf.Name(), w.ipser.Name, "To route")
	} else {
		w.toForward_i("", w.ipser.Name, "To route")
	}
}

func (w *WorkerImpl) newVPN() {
	_, vpn := w.GetCfgs()
	if vpn != nil {
		obj := NewOpenVPN(vpn)
		obj.Initialize()
		w.vpn = obj
	}
}

func (w *WorkerImpl) addVPNRoute(rt co.PrefixRoute) {
	vpn := w.cfg.OpenVPN
	if vpn == nil {
		return
	}

	routes := vpn.Routes
	vpn.Routes = w.setVPNRoute(routes, rt)
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

func (w *WorkerImpl) AddRoute(route *schema.PrefixRoute, v api.SwitchApi) error {
	rt := w.correctRoute(route)
	if !w.cfg.AddRoute(rt) {
		w.out.Info("WorkerImpl.AddRoute: %s route exist", route.Prefix)
		return nil
	}

	w.out.Info("WorkerImpl.AddRoute: %s", rt.String())
	w.addIPSet(rt)
	w.addVPNRoute(rt)
	w.toRoute(rt)
	return nil
}

func (w *WorkerImpl) DelRoute(route *schema.PrefixRoute, v api.SwitchApi) error {
	correctRt := w.correctRoute(route)
	delRt, removed := w.cfg.DelRoute(correctRt)
	if !removed {
		w.out.Info("WorkerImpl.DelRoute: %s not found", route.Prefix)
		return nil
	}

	w.delIPSet(delRt)
	w.delVPNRoute(delRt)
	w.leftRoute(delRt)
	return nil
}

func (w *WorkerImpl) SaveRoute() {
	w.cfg.SaveRoute()
}

func (w *WorkerImpl) Router() api.RouteApi {
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
		Fallback: data.Fallback,
	}
	if !w.cfg.AddOutput(output) {
		w.out.Info("WorkerImple.AddOutput %s already existed", output.Id())
		return
	}
	output.Correct()
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

func (w *WorkerImpl) ZTruster() api.ZTrustApi {
	return w.ztrust
}

func (w *WorkerImpl) Qoser() api.QosApi {
	return w.qos
}

func (w *WorkerImpl) ACLer() api.ACLApi {
	return w.acl
}

func (w *WorkerImpl) FindHoper() api.FindHopApi {
	return w.findhop
}

func (w *WorkerImpl) AddAddress(value string) {
	if w.br != nil {
		w.br.Open(value)
		w.cfg.Bridge.Address = value
		return
	}
	w.out.Info("WorkerImpl.AddAddress notSupport")
}

func (w *WorkerImpl) DelAddress() {
	if w.br != nil {
		w.br.Close()
		w.cfg.Bridge.Address = ""
		return
	}
	w.out.Info("WorkerImpl.AddAddress notSupport")
}

func (w *WorkerImpl) KillVPNClient(name string) error {
	vpn := w.vpn
	if vpn == nil {
		return libol.NewErr("VPN was disabled")
	}

	return vpn.KillClient(name)
}
