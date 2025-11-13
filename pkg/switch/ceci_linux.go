package cswitch

import (
	"os/exec"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

const (
	CeciBin = "/usr/bin/openceci"
)

type CeciWorker struct {
	*WorkerImpl
	spec *co.CeciSpecifies
}

func NewCeciWorker(c *co.Network) *CeciWorker {
	w := &CeciWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	w.spec, _ = c.Specifies.(*co.CeciSpecifies)
	return w
}

func (w *CeciWorker) Initialize() {
	w.out.Info("CeciWorker.Initialize")
}

func (w *CeciWorker) reloadTcp(obj *co.CeciTcp) {
	name := "/var/openlan/ceci/" + obj.Id()
	out, err := libol.CreateFile(name + ".log")
	if err != nil {
		w.out.Warn("CeciWorker.reloadTcp: %s", err)
		return
	}

	libol.MarshalSave(obj, name+".yaml", true)
	libol.Go(func() {
		w.out.Info("CeciWorker.reloadTcp: %s", obj.Id())
		cmd := exec.Command(CeciBin, "-mode", "tcp", "-conf", name+".yaml")
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			w.out.Warn("CeciWorker.reloadTcp: %s", err)
			return
		}
	})
}

func (w *CeciWorker) reloadHttp(obj *co.CeciHttp) {
	name := "/var/openlan/ceci/" + obj.Id()
	out, err := libol.CreateFile(name + ".log")
	if err != nil {
		w.out.Warn("CeciWorker.reloadTcp: %s", err)
		return
	}

	libol.MarshalSave(obj, name+".yaml", true)
	libol.Go(func() {
		w.out.Info("CeciWorker.reloadHttp: %s", obj.Id())
		cmd := exec.Command(CeciBin, "-mode", "http", "-conf", name+".yaml")
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			w.out.Warn("CeciWorker.reloadHttp: %s", err)
			return
		}
	})
}

func (w *CeciWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()
	w.out.Info("CeciWorker.Start")

	for _, obj := range w.spec.Tcp {
		w.reloadTcp(obj)
	}
	for _, obj := range w.spec.Http {
		w.reloadHttp(obj)
	}
}

func (w *CeciWorker) Stop() {
	w.out.Info("CeciWorker.Stop")
}

func (w *CeciWorker) Reload(v api.SwitchApi) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
