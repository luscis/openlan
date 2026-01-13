package access

import (
	"strings"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/network"
)

type Access struct {
	MixAccess
}

func NewAccess(config *config.Access) *Access {
	p := Access{
		MixAccess: NewMixAccess(config),
	}
	return &p
}

func (p *Access) Initialize() {
	w := p.worker
	w.listener.AddAddr = p.AddAddr
	w.listener.DelAddr = p.DelAddr

	p.MixAccess.Initialize()
}

func (p *Access) routeAdd(prefix string) ([]byte, error) {
	network.RouteDel("", prefix, "")
	out, err := network.RouteAdd("", prefix, p.gateway)
	return out, err
}

func (p *Access) AddAddr(addr, gateway string) error {
	if addr == "" {
		return nil
	}

	// add Access-to-Access
	ips := strings.SplitN(addr, "/", 2)
	if gateway == "" {
		gateway = ips[0]
	}

	out, err := network.AddrAdd(p.IfName(), ips[0], gateway)
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, out)
		return err
	}
	p.out.Info("Access.AddAddr: %s via %s", addr, gateway)

	p.addr = addr
	p.gateway = gateway

	// add directly route.
	out, err = p.routeAdd(addr)
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, out)
	}

	p.AddRoute()
	p.Run1()

	return nil
}

func (p *Access) DelAddr(addr string) error {
	// delete directly route.
	out, err := network.RouteDel(p.IfName(), addr, "")
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, out)
	}
	p.out.Info("Access.DelAddr: route %s via %s", addr, p.IfName())

	// delete Access-to-Access
	ip4 := strings.SplitN(addr, "/", 2)[0]
	out, err = network.AddrDel(p.IfName(), ip4)
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, out)
		return err
	}

	p.out.Info("Access.DelAddr: %s", ip4)
	p.addr = ""
	return nil
}

func (p *Access) AddRoute() error {
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
