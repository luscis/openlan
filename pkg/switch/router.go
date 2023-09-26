package _switch

import (
	"net"
	"time"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
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

func (w *RouterWorker) updateVPN() {
	spec := w.spec
	_, vpn := w.GetCfgs()
	if vpn == nil {
		return
	}

	routes := vpn.Routes
	routes = append(routes, vpn.Subnet)
	for _, sub := range spec.Subnets {
		w.out.Info("RouterWorker.updateVPN subnet %s", sub.CIDR)
		routes = append(routes, sub.CIDR)
	}

	w.WorkerImpl.updateVPN(routes)
}

func (w *RouterWorker) Initialize() {
	w.updateVPN()
	w.WorkerImpl.Initialize()

	w.Forward()
	w.forwardVPN()
}

func (w *RouterWorker) LoadRoutes() {
	// install routes
	cfg := w.cfg
	w.out.Debug("RouterWorker.LoadRoute: %v", cfg.Routes)
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		if rt.NextHop == "" && rt.MultiPath == nil {
			// route's next-hop is local not install again.
			continue
		}
		nlrt := nl.Route{Dst: dst}
		for _, hop := range rt.MultiPath {
			nxhe := &nl.NexthopInfo{
				Hops: hop.Weight,
				Gw:   net.ParseIP(hop.NextHop),
			}
			nlrt.MultiPath = append(nlrt.MultiPath, nxhe)
		}
		if rt.MultiPath == nil {
			nlrt.Gw = net.ParseIP(rt.NextHop)
			nlrt.Priority = rt.Metric
		}
		w.out.Debug("RouterWorker.LoadRoute: %s", nlrt.String())
		promise := &libol.Promise{
			First:  time.Second * 2,
			MaxInt: time.Minute,
			MinInt: time.Second * 10,
		}
		rt_c := rt
		promise.Go(func() error {
			if err := nl.RouteReplace(&nlrt); err != nil {
				w.out.Warn("RouterWorker.LoadRoute: %v %s", nlrt, err)
				return err
			}
			w.out.Info("RouterWorker.LoadRoute: %v success", rt_c.String())
			return nil
		})
	}
}

func (w *RouterWorker) UnLoadRoutes() {
	cfg := w.cfg
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nlRt := nl.Route{Dst: dst}
		if rt.MultiPath == nil {
			nlRt.Gw = net.ParseIP(rt.NextHop)
			nlRt.Priority = rt.Metric
		}
		w.out.Debug("RouterWorker.UnLoadRoute: %s", nlRt.String())
		if err := nl.RouteDel(&nlRt); err != nil {
			w.out.Warn("RouterWorker.UnLoadRoute: %s", err)
			continue
		}
		w.out.Info("RouterWorker.UnLoadRoute: %v", rt.String())
	}
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

func (w *RouterWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.LoadRoutes()
	w.WorkerImpl.Start(v)
}

func (w *RouterWorker) Stop() {
	w.WorkerImpl.Stop()
	w.UnLoadRoutes()
}

func (w *RouterWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
