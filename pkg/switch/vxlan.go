package _switch

import (
	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
)

type VxLANWorker struct {
	*WorkerImpl
	spec *co.VxLANSpecifies
}

func NewVxLANWorker(c *co.Network) *VxLANWorker {
	w := &VxLANWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	w.spec, _ = c.Specifies.(*co.VxLANSpecifies)
	return w
}

func (w *VxLANWorker) Initialize() {
	w.WorkerImpl.Initialize()
}

func (w *VxLANWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	master := GetFabricer(w.spec.Fabric)
	if master == nil {
		w.out.Warn("VxLANWorker.Start %s not found", w.spec.Fabric)
		return
	}
	w.cfg.Bridge.Mss = master.TcpMss()
	master.AddNetwork(w.cfg)
	w.WorkerImpl.Start(v)
}

func (w *VxLANWorker) Stop() {
	w.WorkerImpl.Stop()
	master := GetFabricer(w.spec.Fabric)
	if master == nil {
		w.out.Warn("VxLANWorker.Stop %s not found", w.spec.Fabric)
		return
	}
	master.DelNetwork(w.cfg.Bridge.Name, w.spec.Vni)
}

func (w *VxLANWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
