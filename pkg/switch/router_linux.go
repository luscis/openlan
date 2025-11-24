package cswitch

import (
	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	nl "github.com/vishvananda/netlink"
)

type RouterWorker struct {
	*WorkerImpl
	spec      *co.RouterSpecifies
	addresses []*nl.Addr
}

func NewRouterWorker(c *co.Network) *RouterWorker {
	w := &RouterWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	w.spec, _ = c.Specifies.(*co.RouterSpecifies)
	return w
}

func (w *RouterWorker) Initialize() {
	w.WorkerImpl.Initialize()

	spec := w.spec
	if spec.Loopback != "" {
		if addr, err := nl.ParseAddr(spec.Loopback); err == nil {
			w.addresses = append(w.addresses, addr)
		}
	}
	for _, _addr := range spec.Addresses {
		if addr, err := nl.ParseAddr(_addr); err == nil {
			w.addresses = append(w.addresses, addr)
		}
	}

	w.Forward()
}

func (w *RouterWorker) Forward() {
	spec := w.spec
	// Enable MASQUERADE, and FORWARD it.
	w.out.Debug("RouterWorker.Forward %v", w.cfg)
	for _, sub := range spec.Subnets {
		if sub.CIDR == "" {
			continue
		}
		w.setR.Add(sub.CIDR)
	}
	w.toRelated(spec.Link, "Accept related")
	w.toForward_s(spec.Link, w.setR.Name, "", "From route")
	w.toMasq_s(w.setR.Name, "", "To Masq")
}

func (w *RouterWorker) addAddress() error {
	link, err := nl.LinkByName("lo")
	if err != nil {
		w.out.Warn("addAddress: %s", err)
		return err
	}
	for _, addr := range w.addresses {
		if err := nl.AddrAdd(link, addr); err != nil {
			w.out.Warn("addAddress: %s: %s", addr, err)
			continue
		}
		w.out.Info("addAddress %s on lo", addr)
	}
	return nil
}

func (w *RouterWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()

	for _, tun := range w.spec.Tunnels {
		w.AddTunnel(tun)
	}

	w.WorkerImpl.Start(v)
	w.addAddress()
}

func (w *RouterWorker) delAddress() error {
	link, err := nl.LinkByName("lo")
	if err != nil {
		w.out.Warn("delAddress: %s", err)
		return err
	}
	for _, addr := range w.addresses {
		if err := nl.AddrDel(link, addr); err != nil {
			w.out.Warn("delAddress: %s: %s", addr, err)
			continue
		}
		w.out.Info("delAddress %s on lo", addr)
	}
	return nil
}

func (w *RouterWorker) Stop() {
	w.delAddress()
	w.WorkerImpl.Stop()

	for _, tun := range w.spec.Tunnels {
		w.DelTunnel(tun)
	}
}

func (w *RouterWorker) Reload(v api.SwitchApi) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *RouterWorker) AddTunnel(data *co.RouterTunnel) {
	var link nl.Link

	switch data.Protocol {
	case "gre":
		link = &nl.Gretun{
			LinkAttrs: nl.LinkAttrs{
				Name: data.Link,
			},
			Local:  libol.ParseAddr("0.0.0.0"),
			Remote: libol.ParseAddr(data.Remote),
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.AddTunnel.gre %s %s", data.Id(), err)
			return
		}
	case "ipip":
		link = &nl.Iptun{
			LinkAttrs: nl.LinkAttrs{
				Name: data.Link,
			},
			Local:  libol.ParseAddr("0.0.0.0"),
			Remote: libol.ParseAddr(data.Remote),
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.AddTunnel.ip %s %s", data.Id(), err)
			return
		}
	}

	if link == nil {
		return
	}

	addr, err := nl.ParseAddr(data.Address)
	if err == nil {
		if err := nl.AddrAdd(link, addr); err != nil {
			w.out.Warn("WorkerImpl.AddTunnel.addAddr: %s: %s", addr, err)
			return
		}
	}
	if err := nl.LinkSetUp(link); err != nil {
		w.out.Warn("WorkerImpl.AddTunnel.up: %s: %s", data.Id(), err)
	}
}

func (w *RouterWorker) DelTunnel(data *co.RouterTunnel) {
	if data.Link == "" {
		return
	}
	if link, err := nl.LinkByName(data.Link); err == nil {
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.DelTunnel %s %s", data.Id(), err)
			return
		}
	}
}
