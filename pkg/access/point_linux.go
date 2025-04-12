package access

import (
	"net"
	"strings"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
	"github.com/vishvananda/netlink"
)

type Point struct {
	MixPoint
	// private
	brName string
	ipMtu  int
	addr   string
	bypass *netlink.Route
	routes []*models.Route
	link   netlink.Link
	uuid   string
}

func NewPoint(config *config.Point) *Point {
	ipMtu := config.Interface.IPMtu
	if ipMtu == 0 {
		ipMtu = 1500
	}
	p := Point{
		ipMtu:    ipMtu,
		brName:   config.Interface.Bridge,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	w := p.worker

	w.listener.AddAddr = p.AddAddr
	w.listener.DelAddr = p.DelAddr
	w.listener.AddRoutes = p.AddRoutes
	w.listener.DelRoutes = p.DelRoutes
	w.listener.OnTap = p.OnTap
	w.listener.Forward = p.Forward

	p.MixPoint.Initialize()
}

func (p *Point) DelAddr(ipStr string) error {
	if p.link == nil || ipStr == "" {
		return nil
	}
	ipAddr, err := netlink.ParseAddr(ipStr)
	if err != nil {
		p.out.Error("Access.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := netlink.AddrDel(p.link, ipAddr); err != nil {
		p.out.Warn("Access.DelAddr.UnsetLinkIp: %s", err)
	}
	p.out.Info("Access.DelAddr: %s", ipStr)
	p.addr = ""
	return nil
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" || p.link == nil {
		return nil
	}
	ipAddr, err := netlink.ParseAddr(ipStr)
	if err != nil {
		p.out.Error("Access.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := netlink.AddrAdd(p.link, ipAddr); err != nil {
		p.out.Warn("Access.AddAddr.SetLinkIp: %s", err)
		return err
	}
	p.out.Info("Access.AddAddr: %s", ipStr)
	p.addr = ipStr
	return nil
}

func (p *Point) UpBr(name string) *netlink.Bridge {
	if name == "" {
		return nil
	}
	la := netlink.LinkAttrs{TxQLen: -1, Name: name}
	br := &netlink.Bridge{LinkAttrs: la}
	if link, err := netlink.LinkByName(name); link == nil {
		p.out.Warn("Access.UpBr: %s %s", name, err)
		err := netlink.LinkAdd(br)
		if err != nil {
			p.out.Warn("Access.UpBr.newBr: %s %s", name, err)
		}
	}
	link, err := netlink.LinkByName(name)
	if link == nil {
		p.out.Error("Access.UpBr: %s %s", name, err)
		return nil
	}
	if err := netlink.LinkSetUp(link); err != nil {
		p.out.Error("Access.UpBr.LinkUp: %s", err)
	}
	return br
}

func (p *Point) OnTap(w *TapWorker) error {
	p.out.Info("Access.OnTap")
	tap := w.device
	name := tap.Name()
	if tap.Type() == network.ProviderVir { // virtual device
		p.out.Error("Access.OnTap: not support %s", tap.Type())
		return nil
	}
	// kernel device
	link, err := netlink.LinkByName(name)
	if err != nil {
		p.out.Error("Access.OnTap: Get %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetMTU(link, p.ipMtu); err != nil {
		p.out.Error("Access.OnTap.SetMTU: %s", err)
	}
	if br := p.UpBr(p.brName); br != nil {
		if err := netlink.LinkSetMaster(link, br); err != nil {
			p.out.Error("Access.OnTap.AddSlave: Switch dev %s: %s", name, err)
		}
		link, err = netlink.LinkByName(p.brName)
		if err != nil {
			p.out.Error("Access.OnTap: Get %s: %s", p.brName, err)
		}
		if p.config.Interface.Cost > 0 {
			port := network.NewBrPort(name)
			if err := port.Cost(p.config.Interface.Cost); err != nil {
				p.out.Error("Access.OnTap: Cost %s: %s", p.brName, err)
			}
		}
	}
	p.link = link
	return nil
}

func (p *Point) GetRemote() string {
	conn := p.worker.conWorker
	if conn == nil {
		return ""
	}
	remote := conn.client.RemoteAddr()
	remote = strings.SplitN(remote, ":", 2)[0]
	return remote
}

func (p *Point) AddBypass(routes []*models.Route) {
	remote := p.GetRemote()
	if !p.config.ByPass {
		return
	}
	addr, dest, _ := net.ParseCIDR(remote + "/32")
	gws, err := netlink.RouteGet(addr)
	if err != nil || len(gws) == 0 {
		p.out.Error("Access.AddBypass: RouteGet %s: %s", addr, err)
		return
	}
	rt := &netlink.Route{
		LinkIndex: gws[0].LinkIndex,
		Dst:       dest,
		Gw:        gws[0].Gw,
		Table:     100,
	}
	p.out.Debug("Access.AddBypass: %s")
	if err := netlink.RouteReplace(rt); err != nil {
		p.out.Warn("Access.AddBypass: %s %s", rt.Dst, err)
		return
	}
	p.out.Info("Access.AddBypass: route %s via %s", rt.Dst, rt.Gw)
	ru := netlink.NewRule()
	ru.Table = 100
	ru.Priority = 16383
	if err := netlink.RuleAdd(ru); err != nil {
		p.out.Warn("Access.AddBypass: %s %s", ru.Dst, err)
	}
	p.out.Info("Access.AddBypass: %s", ru)
	p.bypass = rt
	for _, rt := range routes {
		if rt.Prefix != "0.0.0.0/0" {
			continue
		}
		gw := net.ParseIP(rt.NextHop)
		_, dst0, _ := net.ParseCIDR("0.0.0.0/1")
		rt0 := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst0,
			Gw:        gw,
			Priority:  rt.Metric,
		}
		p.out.Debug("Access.AddBypass: %s", rt0)
		if err := netlink.RouteAdd(&rt0); err != nil {
			p.out.Warn("Access.AddBypass: %s %s", rt0.Dst, err)
		}
		p.out.Info("Access.AddBypass: route %s via %s", rt0.Dst, rt0.Gw)
		_, dst1, _ := net.ParseCIDR("128.0.0.0/1")
		rt1 := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst1,
			Gw:        gw,
			Priority:  rt.Metric,
		}
		p.out.Debug("Access.AddBypass: %s", rt1)
		if err := netlink.RouteAdd(&rt1); err != nil {
			p.out.Warn("Access.AddBypass: %s %s", rt1.Dst, err)
		}
		p.out.Info("Access.AddBypass: route %s via %s", rt1.Dst, rt1.Gw)
	}
}

func (p *Point) Forward(prefix, nexthop string) {
	_, dst, _ := net.ParseCIDR(prefix)
	p.out.Info("Access.Forward: %s via %s", prefix, nexthop)
	if err := netlink.RouteAdd(&netlink.Route{
		Dst: dst,
		Gw:  net.ParseIP(nexthop),
	}); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			p.out.Warn("Access.Forward: %s %s", prefix, err)
		}
	}
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil || p.link == nil {
		return nil
	}
	p.AddBypass(routes)
	for _, rt := range routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(rt.NextHop)
		rte := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst,
			Gw:        nxt,
			Priority:  rt.Metric,
		}
		p.out.Debug("Access.AddRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			p.out.Warn("Access.AddRoute: %s %s", rt.Prefix, err)
			continue
		}
		p.out.Info("Access.AddRoutes: route %s via %s", rt.Prefix, rt.NextHop)
	}
	p.routes = routes
	return nil
}

func (p *Point) DelBypass(routes []*models.Route) {
	if !p.config.ByPass || p.bypass == nil {
		return
	}
	p.out.Debug("Access.DelRoute: %s")
	rt := p.bypass
	if err := netlink.RouteAdd(rt); err != nil {
		p.out.Warn("Access.DelRoute: %s %s", rt.Dst, err)
	}
	p.out.Info("Access.DelBypass: route %s via %s", rt.Dst, rt.Gw)
	p.bypass = nil
	for _, rt := range routes {
		if rt.Prefix != "0.0.0.0/0" {
			continue
		}
		gw := net.ParseIP(rt.NextHop)
		_, dst0, _ := net.ParseCIDR("0.0.0.0/1")
		rt0 := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst0,
			Gw:        gw,
			Priority:  rt.Metric,
		}
		p.out.Debug("Access.DelBypass: %s", rt0)
		if err := netlink.RouteDel(&rt0); err != nil {
			p.out.Warn("Access.DelBypass: %s %s", rt0.Dst, err)
		}
		p.out.Info("Access.DelBypass: route %s via %s", rt0.Dst, rt0.Gw)
		_, dst1, _ := net.ParseCIDR("128.0.0.0/1")
		rt1 := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst1,
			Gw:        gw,
			Priority:  rt.Metric,
		}
		p.out.Debug("Access.DelBypass: %s", rt1)
		if err := netlink.RouteDel(&rt1); err != nil {
			p.out.Warn("Access.DelBypass: %s %s", rt1.Dst, err)
		}
		p.out.Info("Access.DelBypass: route %s via %s", rt1.Dst, rt1.Gw)
	}
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	if routes == nil || p.link == nil {
		return nil
	}
	p.DelBypass(routes)
	for _, rt := range routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(rt.NextHop)
		rte := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst,
			Gw:        nxt,
			Priority:  rt.Metric,
		}
		if err := netlink.RouteDel(&rte); err != nil {
			p.out.Warn("Access.DelRoute: %s %s", rt.Prefix, err)
			continue
		}
		p.out.Info("Access.DelRoutes: route %s via %s", rt.Prefix, rt.NextHop)
	}
	p.routes = nil
	return nil
}
