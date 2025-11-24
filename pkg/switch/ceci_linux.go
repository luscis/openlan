package cswitch

import (
	"os"
	"os/exec"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

const (
	CeciBin = "/usr/bin/openceci"
	CeciDir = "/var/openlan/ceci/"
)

type CeciWorker struct {
	*WorkerImpl
	spec *co.CeciSpecifies
}

func NewCeciWorker(c *co.Network) *CeciWorker {
	w := &CeciWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	api.Call.SetCeciApi(w)
	w.spec, _ = c.Specifies.(*co.CeciSpecifies)
	return w
}

func (w *CeciWorker) Initialize() {
	w.out.Info("CeciWorker.Initialize")
	w.addCache()
}

func (w *CeciWorker) killPid(name string) {
	if libol.FileExist(name) != nil {
		return
	}

	pid, _ := os.ReadFile(name)
	kill, _ := exec.LookPath("kill")
	cmd := exec.Command(kill, string(pid))
	if err := cmd.Run(); err != nil {
		w.out.Warn("CeciWorker.killPid:%s: %s", pid, err)
		return
	}
}

func (w *CeciWorker) reloadTcp(obj *co.CeciTcp) {
	name := CeciDir + obj.Id()
	out, err := libol.CreateFile(name + ".log")
	if err != nil {
		w.out.Warn("CeciWorker.reloadTcp: %s", err)
		return
	}

	w.killPid(name + ".pid")
	libol.MarshalSave(obj, name+".yaml", true)
	libol.Go(func() {
		w.out.Info("CeciWorker.reloadTcp: %s", obj.Id())
		cmd := exec.Command(CeciBin, "-mode", obj.Mode, "-conf", name+".yaml", "-write-pid", name+".pid")
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			w.out.Warn("CeciWorker.reloadTcp: %s", err)
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
}

func (w *CeciWorker) Stop() {
	w.out.Info("CeciWorker.Stop")
}

func (w *CeciWorker) Reload(v api.SwitchApi) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *CeciWorker) AddTcp(data schema.CeciTcp) {
	obj := &co.CeciTcp{
		Mode:   data.Mode,
		Listen: data.Listen,
		Target: data.Target,
	}

	obj.Correct()
	if t, _ := w.spec.FindTcp(obj); t == nil {
		w.spec.AddTcp(obj)
	} else {
		t.Target = data.Target
	}

	w.reloadTcp(obj)
}

func (w *CeciWorker) DelTcp(data schema.CeciTcp) {
	obj := &co.CeciTcp{
		Listen: data.Listen,
	}
	obj.Correct()
	if _, removed := w.spec.DelTcp(obj); removed {
		name := CeciDir + obj.Id()
		w.killPid(name + ".pid")
	}
}
