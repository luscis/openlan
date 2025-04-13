package cswitch

import (
	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	nl "github.com/vishvananda/netlink"
)

type RouterWorker struct {
	*WorkerImpl
	spec *co.RouterSpecifies
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
	spec := w.spec
	if spec.Loopback == "" {
		return nil
	}
	link, err := nl.LinkByName("lo")
	if err != nil {
		w.out.Warn("addAddress: %s", err)
		return err
	}
	addr, err := nl.ParseAddr(spec.Loopback)
	if err != nil {
		w.out.Warn("addAddress: %s", err)
		return err
	}
	if err := nl.AddrAdd(link, addr); err != nil {
		w.out.Warn("addAddress: %s", err)
		return err
	}
	return nil
}

func (w *RouterWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.WorkerImpl.Start(v)
	w.addAddress()
}

func (w *RouterWorker) delAddress() error {
	spec := w.spec
	if spec.Loopback == "" {
		return nil
	}
	link, err := nl.LinkByName("lo")
	if err != nil {
		w.out.Warn("delAddress: %s", err)
		return err
	}
	addr, err := nl.ParseAddr(spec.Loopback)
	if err != nil {
		w.out.Warn("delAddress: %s", err)
		return err
	}
	if err := nl.AddrDel(link, addr); err != nil {
		w.out.Warn("delAddress: %s", err)
		return err
	}
	return nil
}

func (w *RouterWorker) Stop() {
	w.delAddress()
	w.WorkerImpl.Stop()
}

func (w *RouterWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
