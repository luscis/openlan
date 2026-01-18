package cswitch

import (
	"net"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

type RouterWorker struct {
	*WorkerImpl
	spec      *co.RouterSpecifies
	addresses []*nl.Addr
	ipses     *cn.IPSet
}

func NewRouterWorker(c *co.Network) *RouterWorker {
	w := &RouterWorker{
		WorkerImpl: NewWorkerApi(c),
		ipses:      cn.NewIPSet(c.Name+"_s", "hash:net"),
	}
	api.Call.SetRouterApi(w)
	w.spec, _ = c.Specifies.(*co.RouterSpecifies)
	return w
}

func (w *RouterWorker) Initialize() {
	w.newFire()
	w.ipses.Clear()
	w.addCache()
}

func (w *RouterWorker) SourceNAT() {
	spec := w.spec
	// Enable MASQUERADE, and FORWARD it.
	w.out.Debug("RouterWorker.SourceNAT %v", w.cfg)
	for _, sub := range spec.Private {
		w.ipses.Add(sub)
	}
	w.toForward_s("", w.ipses.Name, "", "From route")
	w.toMasq_s(w.ipses.Name, "", "To Masq")
}

func (w *RouterWorker) addAddress(name string, addrs []string) error {
	link, err := nl.LinkByName(name)
	if err != nil {
		w.out.Warn("RouterWorker.addAddress: %s", err)
		return err
	}
	for _, addr := range addrs {
		if _addr, err := nl.ParseAddr(addr); err == nil {
			if err := nl.AddrAdd(link, _addr); err != nil {
				w.out.Warn("RouterWorker.addAddress: %s: %s", addr, err)
				continue
			}
			w.out.Info("RouterWorker.addAddress %s on %s", addr, name)
		}
	}
	return nil
}

func (w *RouterWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()
	w.fire.Start()
	w.toSNAT()

	for _, tun := range w.spec.Tunnels {
		w.addTunnel(tun)
	}

	w.SourceNAT()
	if w.spec.Loopback != "" {
		w.addAddress("lo", []string{w.spec.Loopback})
	}
	w.addAddress("lo", w.spec.Addresses)

	for _, port := range w.spec.Interfaces {
		w.addInterface(port)
	}
	for _, re := range w.spec.Redirect {
		w.addRedirect(re)
	}
}

func (w *RouterWorker) delAddress(name string, addrs []string) error {
	link, err := nl.LinkByName(name)
	if err != nil {
		w.out.Warn("RouterWorker.delAddress: %s", err)
		return err
	}
	for _, addr := range addrs {
		if _addr, err := nl.ParseAddr(addr); err == nil {
			if err := nl.AddrDel(link, _addr); err != nil {
				w.out.Warn("RouterWorker.delAddress: %s: %s", addr, err)
				continue
			}
			w.out.Info("RouterWorker.delAddress %s on %s", addr, name)
		}
	}
	return nil
}

func (w *RouterWorker) Stop(kill bool) {
	if kill {
		for _, re := range w.spec.Redirect {
			w.delRedirect(re)
		}
		for _, port := range w.spec.Interfaces {
			w.delInterface(port)
		}
		if w.spec.Loopback != "" {
			w.delAddress("lo", []string{w.spec.Loopback})
		}
		w.delAddress("lo", w.spec.Addresses)
		for _, tun := range w.spec.Tunnels {
			w.delTunnel(tun)
		}
		w.ipses.Destroy()
		w.fire.Stop()
	}
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
		if li, _ := nl.LinkByName(data.Link); li != nil {
			w.out.Warn("RouterWorker.addTunnel %s existed", data.Link)
			nl.LinkDel(li)
		}
		if err := nl.LinkAdd(link); err != nil {
			w.out.Error("RouterWorker.AddTunnel.gre %s %s", data.ID(), err)
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
			w.out.Error("RouterWorker.AddTunnel.ip %s %s", data.ID(), err)
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
		w.out.Warn("RouterWorker.AddTunnel.up: %s: %s", data.ID(), err)
	}
}

func (w *RouterWorker) delTunnel(data *co.RouterTunnel) {
	if link, err := nl.LinkByName(data.Link); err == nil {
		if err := nl.LinkDel(link); err != nil {
			w.out.Error("RouterWorker.DelTunnel %s %s", data.ID(), err)
			return
		}
	} else {
		w.out.Warn("RouterWorker.DelTunnel notFound %s:%s", data.ID(), data.Link)
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

func (w *RouterWorker) AddPrivate(data string) error {
	if ok := w.spec.AddPrivate(data); ok {
		w.ipses.Add(data)
	}
	return nil
}

func (w *RouterWorker) DelPrivate(data string) error {
	if old, ok := w.spec.DelPrivate(data); ok {
		w.ipses.Del(old)
	}
	return nil
}

func (w *RouterWorker) addInterface(value *co.RouterInterface) {
	link, err := nl.LinkByName(value.Device)
	if link == nil {
		w.out.Error("RouterWorker.addInterface %s %s", value.Device, err)
		return
	}
	if err := nl.LinkSetUp(link); err != nil {
		w.out.Warn("RouterWorker.addInterface %s %s", value.Device, err)
	}

	name := value.ID()
	if value.VLAN > 0 {
		vLink := &nl.Vlan{
			LinkAttrs: nl.LinkAttrs{
				Name:        value.ID(),
				ParentIndex: link.Attrs().Index,
			},
			VlanId: value.VLAN,
		}
		if li, _ := nl.LinkByName(name); li != nil {
			w.out.Warn("RouterWorker.addInterface %s existed", name)
			nl.LinkDel(li)
		}
		if err := nl.LinkAdd(vLink); err != nil {
			w.out.Error("RouterWorker.addInterface %s %s", name, err)
			return
		}
		if err := nl.LinkSetUp(vLink); err != nil {
			w.out.Warn("RouterWorker.addInterface %s %s", name, err)
		}
	}
	w.addAddress(name, []string{value.Address})
}

func (w *RouterWorker) AddInterface(data schema.RouterInterface) error {
	obj := &co.RouterInterface{
		Device:  data.Device,
		VLAN:    data.VLAN,
		Address: data.Address,
	}
	if ok := w.spec.AddInterface(obj); ok {
		w.addInterface(obj)
	}
	return nil
}

func (w *RouterWorker) delInterface(value *co.RouterInterface) {
	name := value.ID()
	link, err := nl.LinkByName(name)
	if link == nil {
		w.out.Error("RouterWorker.delInterface %s %s", value.Device, err)
		return
	}

	w.delAddress(name, []string{value.Address})
	if value.VLAN > 0 {
		if err := nl.LinkDel(link); err != nil {
			return
		}
	}
}

func (w *RouterWorker) DelInterface(data schema.RouterInterface) error {
	obj := &co.RouterInterface{
		Device: data.Device,
		VLAN:   data.VLAN,
	}
	if old, ok := w.spec.DelInterface(obj); ok {
		w.delInterface(old)
	}
	return nil
}

func (w *RouterWorker) addRedirect(obj *co.RouterRedirect) {
	route := nl.Route{
		Table: obj.Table,
		Gw:    net.ParseIP(obj.NextHop),
	}
	route.Dst, _ = libol.ParseNet("0.0.0.0/0")
	if route.Gw == nil {
		w.out.Warn("WorkerImpl.AddRedirect: invalid %s", obj.NextHop)
		return
	}
	promise := libol.NewPromise()
	promise.Go(func() error {
		if err := nl.RouteReplace(&route); err != nil {
			w.out.Warn("WorkerImpl.AddRedirect: %s %s", route, err)
			return err
		}
		w.out.Info("WorkerImpl.AddRedirect: %s success", obj.Route())
		if out, err := cn.RuleAdd(obj.Source, obj.Table, obj.Priority); err != nil {
			w.out.Warn("WorkerImpl.AddRedirect: %s %s", obj.Rule(), out)
		} else {
			w.out.Info("WorkerImpl.AddRedirect: %s success", obj.Rule())
		}
		return nil
	})
}

func (w *RouterWorker) AddRedirect(value schema.RedirectRoute) {
	obj := &co.RouterRedirect{
		Source:  value.Source,
		Table:   value.Table,
		NextHop: value.NextHop,
	}
	obj.Correct()
	if ok := w.spec.AddRedirect(obj); ok {
		w.addRedirect(obj)
	}
}

func (w *RouterWorker) delRedirect(obj *co.RouterRedirect) {
	prefix, _ := libol.ParseNet("0.0.0.0/0")
	route := nl.Route{Dst: prefix, Table: obj.Table}
	if err := nl.RouteDel(&route); err != nil {
		w.out.Warn("WorkerImpl.DelRedirect: %v %s", route, err)
	} else {
		w.out.Info("WorkerImpl.DelRedirect: %v success", route.String())
	}

	if _, err := cn.RuleDel(obj.Source, obj.Table); err != nil {
		w.out.Warn("WorkerImpl.DelRedirect: %s %s", obj.Rule(), err)
	} else {
		w.out.Info("WorkerImpl.DelRedirect: %v success", obj.Rule())
	}
}

func (w *RouterWorker) DelRedirect(value schema.RedirectRoute) {
	obj := &co.RouterRedirect{
		Source:  value.Source,
		Table:   value.Table,
		NextHop: value.NextHop,
	}
	obj.Correct()
	if old, ok := w.spec.DelRedirect(obj); ok {
		w.delRedirect(old)
	}
}
