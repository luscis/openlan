package access

import (
	"strings"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/network"
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

	p.MixPoint.Initialize()
}

func (p *Point) routeAdd(prefix string) ([]byte, error) {
	network.RouteDel("", prefix, "")
	out, err := network.RouteAdd(p.IfName(), prefix, "")
	return out, err
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}

	// add point-to-point
	ips := strings.SplitN(ipStr, "/", 2)
	out, err := network.AddrAdd(p.IfName(), ips[0], ips[0])
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, out)
		return err
	}
	p.out.Info("Access.AddAddr: %s", ipStr)

	// add directly route.
	out, err = p.routeAdd(ipStr)
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, out)
	}
	p.AddRoute()

	p.addr = ipStr

	return nil
}

func (p *Point) DelAddr(ipStr string) error {
	// delete directly route.
	out, err := network.RouteDel(p.IfName(), ipStr, "")
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, out)
	}
	p.out.Info("Access.DelAddr: route %s via %s", ipStr, p.IfName())

	// delete point-to-point
	ip4 := strings.SplitN(ipStr, "/", 2)[0]
	out, err = network.AddrDel(p.IfName(), ip4)
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, out)
		return err
	}

	p.out.Info("Access.DelAddr: %s", ip4)
	p.addr = ""
	return nil
}

func (p *Point) AddRoute() error {
	to := p.config.Forward
	if to == nil {
		return nil
	}

	for _, prefix := range to {
		out, err := p.routeAdd(prefix)
		if err != nil {
			p.out.Warn("Access.AddRoute: %s: %s", prefix, out)
			continue
		}
		p.out.Info("Access.AddRoute: %s via %s", prefix, p.IfName())
	}
	return nil
}
