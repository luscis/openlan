package access

import (
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"strings"
)

type Point struct {
	MixPoint
	// private
	brName string
	addr   string
	routes []*models.Route
}

func NewPoint(config *config.Point) *Point {
	p := Point{
		brName:   config.Interface.Bridge,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	p.worker.listener.AddAddr = p.AddAddr
	p.worker.listener.DelAddr = p.DelAddr
	p.worker.listener.AddRoutes = p.AddRoutes
	p.worker.listener.DelRoutes = p.DelRoutes
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
		p.out.Warn("Point.AddAddr: %s, %s", err, out)
		return err
	}
	p.out.Info("Point.AddAddr: %s", ipStr)
	// add directly route.
	out, err = libol.IpRouteAdd(p.IfName(), ipStr, "")
	if err != nil {
		p.out.Warn("Point.AddAddr: %s, %s", err, out)
	}
	p.out.Info("Point.AddAddr: route %s via %s", ipStr, p.IfName())
	p.addr = ipStr
	return nil
}

func (p *Point) DelAddr(ipStr string) error {
	// delete directly route.
	out, err := libol.IpRouteDel(p.IfName(), ipStr, "")
	if err != nil {
		p.out.Warn("Point.DelAddr: %s, %s", err, out)
	}
	p.out.Info("Point.DelAddr: route %s via %s", ipStr, p.IfName())
	// delete point-to-point
	ip4 := strings.SplitN(ipStr, "/", 2)[0]
	out, err = libol.IpAddrDel(p.IfName(), ip4)
	if err != nil {
		p.out.Warn("Point.DelAddr: %s, %s", err, out)
		return err
	}
	p.out.Info("Point.DelAddr: %s", ip4)
	p.addr = ""
	return nil
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil {
		return nil
	}
	for _, route := range routes {
		out, err := libol.IpRouteAdd(p.IfName(), route.Prefix, "")
		if err != nil {
			p.out.Warn("Point.AddRoutes: %s %s", route.Prefix, out)
			continue
		}
		p.out.Info("Point.AddRoutes: route %s via %s", route.Prefix, p.IfName())
	}
	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	for _, route := range routes {
		out, err := libol.IpRouteDel(p.IfName(), route.Prefix, "")
		if err != nil {
			p.out.Warn("Point.DelRoutes: %s %s", route.Prefix, out)
			continue
		}
		p.out.Info("Point.DelRoutes: route %s via %s", route.Prefix, p.IfName())
	}
	p.routes = nil
	return nil
}
