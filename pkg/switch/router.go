package cswitch

import (
	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
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

func (w *RouterWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.WorkerImpl.Start(v)
}

func (w *RouterWorker) Stop() {
	w.WorkerImpl.Stop()
}

func (w *RouterWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
