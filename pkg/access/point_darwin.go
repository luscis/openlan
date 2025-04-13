package access

import (
	"strings"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

type Point struct {
	MixPoint
	// private
	brName string
	addr   string
}

func NewPoint(config *config.Point) *Point {
	p := Point{
		brName:   config.Interface.Bridge,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	w := p.worker
	w.listener.AddAddr = p.AddAddr
	w.listener.DelAddr = p.DelAddr
	w.listener.Forward = p.Forward

	p.MixPoint.Initialize()
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}
	// add point-to-point
	ips := strings.SplitN(ipStr, "/", 2)
	out, err := libol.IpAddrAdd(p.IfName(), ips[0], ips[0])
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, out)
		return err
	}
	p.out.Info("Access.AddAddr: %s", ipStr)
	// add directly route.
	out, err = libol.IpRouteAdd(p.IfName(), ipStr, "")
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, out)
	}
	p.out.Info("Access.AddAddr: route %s via %s", ipStr, p.IfName())
	p.addr = ipStr

	p.AddRoutes()

	return nil
}

func (p *Point) DelAddr(ipStr string) error {
	// delete directly route.
	out, err := libol.IpRouteDel(p.IfName(), ipStr, "")
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, out)
	}
	p.out.Info("Access.DelAddr: route %s via %s", ipStr, p.IfName())
	// delete point-to-point
	ip4 := strings.SplitN(ipStr, "/", 2)[0]
	out, err = libol.IpAddrDel(p.IfName(), ip4)
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, out)
		return err
	}
	p.out.Info("Access.DelAddr: %s", ip4)
	p.addr = ""
	return nil
}

func (p *Point) AddRoutes() error {
	to := p.config.Forward
	if to == nil {
		return nil
	}

	for _, prefix := range to.Match {
		out, err := libol.IpRouteAdd("", prefix, to.Server)
		if err != nil {
			p.out.Warn("Access.AddRoutes: %s %s", prefix, out)
			continue
		}
		p.out.Info("Access.AddRoutes: route %s via %s", prefix, to.Server)
	}
	return nil
}

func (p *Point) Forward(name, prefix, nexthop string) {
	if out, err := libol.IpRouteAdd("", prefix, nexthop); err != nil {
		if strings.Contains(err.Error(), "file exists") {
			return
		}
		p.out.Warn("Access.Forward: %s %s: %s", prefix, err, out)
		return
	}
	p.out.Info("Access.Forward: %s %s via %s", name, prefix, nexthop)
}
