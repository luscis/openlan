package access

import (
	"strings"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
)

type Access struct {
	MixAccess
	// private
	routes []*models.Route
	config *config.Access
}

func NewAccess(config *config.Access) *Access {
	p := Access{
		MixAccess: NewMixAccess(config),
	}
	return &p
}

func (p *Access) Initialize() {
	p.worker.listener.AddAddr = p.AddAddr
	p.worker.listener.DelAddr = p.DelAddr
	p.worker.listener.OnTap = p.OnTap
	p.MixAccess.Initialize()
}

func (p *Access) OnTap(w *TapWorker) error {
	// clean routes previous
	routes := make([]*models.Route, 0, 32)
	if err := libol.UnmarshalLoad(&routes, ".routes.json"); err == nil {
		for _, route := range routes {
			_, _ = network.RouteDel(p.IfName(), route.Prefix, route.NextHop)
			p.out.Debug("Access.OnTap: clear %s via %s", route.Prefix, route.NextHop)
		}
	}
	return nil
}

func (p *Access) Trim(out []byte) string {
	return strings.TrimSpace(string(out))
}

func (p *Access) AddAddr(addr, gateway string) error {
	if addr == "" {
		return nil
	}
	addrExisted := network.AddrShow(p.IfName())
	if len(addrExisted) > 0 {
		for _, addr := range addrExisted {
			_, _ = network.AddrDel(p.IfName(), addr)
		}
	}
	out, err := network.AddrAdd(p.IfName(), addr)
	if err != nil {
		p.out.Warn("Access.AddAddr: %s, %s", err, p.Trim(out))
		return err
	}
	p.out.Info("Access.AddAddr: %s", addr)
	p.addr = addr
	p.Run1()

	return nil
}

func (p *Access) DelAddr(ipStr string) error {
	ipv4 := strings.Split(ipStr, "/")[0]
	out, err := network.AddrDel(p.IfName(), ipv4)
	if err != nil {
		p.out.Warn("Access.DelAddr: %s, %s", err, p.Trim(out))
		return err
	}
	p.out.Info("Access.DelAddr: %s", ipv4)
	p.addr = ""
	return nil
}
