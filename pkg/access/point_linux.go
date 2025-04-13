package access

import (
	"net"
	"strings"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
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

	p.AddRoutes()

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

func (p *Point) AddRoutes() error {
	to := p.config.Forward
	if to == nil {
		return nil
	}

	for _, prefix := range to.Match {
		dst, err := libol.ParseCIDR(prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(to.Server)
		rte := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst,
			Gw:        nxt,
		}
		p.out.Debug("Access.AddRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			p.out.Warn("Access.AddRoute: %s %s", prefix, err)
			continue
		}
		p.out.Info("Access.AddRoutes: route %s via %s", prefix, to.Server)
	}
	return nil
}

func (p *Point) Forward(name, prefix, nexthop string) {
	dst, _ := libol.ParseCIDR(prefix)
	if err := netlink.RouteAdd(&netlink.Route{
		Dst: dst,
		Gw:  net.ParseIP(nexthop),
	}); err != nil {
		if strings.Contains(err.Error(), "file exists") {
			return
		}
		p.out.Warn("Access.Forward: %s %s", prefix, err)
		return
	}
	p.out.Info("Access.Forward: %s <- %s %s ", nexthop, name, prefix)
}
