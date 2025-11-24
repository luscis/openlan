package cswitch

import (
	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
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
	api.Call.SetRouterApi(w)
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
		w.out.Warn("RouterWorker.addAddress: %s", err)
		return err
	}
	for _, addr := range w.addresses {
		if err := nl.AddrAdd(link, addr); err != nil {
			w.out.Warn("RouterWorker.addAddress: %s: %s", addr, err)
			continue
		}
		w.out.Info("RouterWorker.addAddress %s on lo", addr)
	}
	return nil
}

func (w *RouterWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()

	for _, tun := range w.spec.Tunnels {
		w.addTunnel(tun)
	}

	w.WorkerImpl.Start(v)
	w.addAddress()
}

func (w *RouterWorker) delAddress() error {
	link, err := nl.LinkByName("lo")
	if err != nil {
		w.out.Warn("RouterWorker.delAddress: %s", err)
		return err
	}
	for _, addr := range w.addresses {
		if err := nl.AddrDel(link, addr); err != nil {
			w.out.Warn("RouterWorker.delAddress: %s: %s", addr, err)
			continue
		}
		w.out.Info("RouterWorker.delAddress %s on lo", addr)
	}
	return nil
}

func (w *RouterWorker) Stop() {
	w.delAddress()
	w.WorkerImpl.Stop()

	for _, tun := range w.spec.Tunnels {
		w.delTunnel(tun)
	}
}

func (w *RouterWorker) Reload(v api.SwitchApi) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *RouterWorker) addTunnel(data *co.RouterTunnel) {
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
			w.out.Error("RouterWorker.AddTunnel.gre %s %s", data.Id(), err)
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
			w.out.Error("RouterWorker.AddTunnel.ip %s %s", data.Id(), err)
			return
		}
	}

	if link == nil {
		return
	}

	addr, err := nl.ParseAddr(data.Address)
	if err == nil {
		if err := nl.AddrAdd(link, addr); err != nil {
			w.out.Warn("RouterWorker.AddTunnel.addAddr: %s: %s", addr, err)
			return
		}

	}
	if err := nl.LinkSetUp(link); err != nil {
		w.out.Warn("RouterWorker.AddTunnel.up: %s: %s", data.Id(), err)
	}
}

func (w *RouterWorker) delTunnel(data *co.RouterTunnel) {
	if link, err := nl.LinkByName(data.Link); err == nil {
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("RouterWorker.DelTunnel %s %s", data.Id(), err)
			return
		}
	} else {
		w.out.Warn("RouterWorker.DelTunnel notFound %s:%s", data.Id(), data.Link)
	}
}

func (w *RouterWorker) AddTunnel(data schema.RouterTunnel) error {
	obj := &co.RouterTunnel{
		Remote:   data.Remote,
		Protocol: data.Protocol,
		Address:  data.Address,
	}
	obj.Correct()
	if ok := w.spec.AddTunnel(obj); ok {
		w.addTunnel(obj)
	}
	return nil
}

func (w *RouterWorker) DelTunnel(data schema.RouterTunnel) error {
	obj := &co.RouterTunnel{
		Remote:   data.Remote,
		Protocol: data.Protocol,
	}
	obj.Correct()
	if old, ok := w.spec.DelTunnel(obj); ok {
		w.delTunnel(old)
	}
	return nil
}
