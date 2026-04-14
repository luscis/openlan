package cswitch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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
	cmd := exec.Command(kill, strings.TrimSpace(string(pid)))
	if err := cmd.Run(); err != nil {
		w.out.Warn("CeciWorker.killPid:%s: %s", pid, err)
		return
	}
}

func (w *CeciWorker) findPid(name string) int {
	pid := 0
	data, err := os.ReadFile(name)
	if err != nil {
		return 0
	}
	if n, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
		pid = n
	}
	return pid
}

func (w *CeciWorker) isRunning(name string) bool {
	pid := w.findPid(name)
	return pid > 0 && libol.HasProcess(pid)
}

func (w *CeciWorker) start(obj *co.CeciProxy) {
	name := CeciDir + obj.Id()

	var data any
	switch obj.Mode {
	case "http":
		data = &co.HttpProxy{
			Listen:   obj.Listen,
			Network:  obj.Network,
			Backends: obj.Backends,
			Cert:     obj.Cert,
		}
	default:
		data = &co.TcpProxy{
			Listen: obj.Listen,
			Target: obj.Target,
		}
	}
	if err := libol.MarshalSave(data, name+".yaml", true); err != nil {
		w.out.Warn("CeciWorker.start: %s", err)
		return
	}

	libol.Go(func() {
		w.out.Info("CeciWorker.start: %s", obj.Id())
		args := []string{"-mode", obj.Mode, "-conf", name + ".yaml"}
		if obj.Network != "" {
			args = append(args, "-network", obj.Network)
		}
		args = append(args, "-log:file", name+".log")
		args = append(args, "-write-pid", name+".pid")
		cmd := exec.Command(CeciBin, args...)
		if err := cmd.Run(); err != nil {
			w.out.Warn("CeciWorker.start: %s", err)
			return
		}
	})
}

func (w *CeciWorker) restart(obj *co.CeciProxy) {
	name := CeciDir + obj.Id()
	w.killPid(name + ".pid")
	w.start(obj)
}

func (w *CeciWorker) saveCertFile(name string, data string) (string, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return "", nil
	}
	if err := os.MkdirAll(CeciDir, 0700); err != nil {
		return "", err
	}
	file := filepath.Join(CeciDir, name)
	if err := os.WriteFile(file, []byte(data+"\n"), 0600); err != nil {
		return "", err
	}
	return file, nil
}

func (w *CeciWorker) applyCertData(obj *co.CeciProxy, data *schema.Cert) error {
	if data == nil {
		return nil
	}

	cert := &co.Cert{
		Insecure: data.Insecure,
	}

	base := obj.Id()
	if data.CrtData != "" || data.KeyData != "" || data.CaData != "" {
		crtFile, err := w.saveCertFile(base+".crt", data.CrtData)
		if err != nil {
			return err
		}
		keyFile, err := w.saveCertFile(base+".key", data.KeyData)
		if err != nil {
			return err
		}
		caFile, err := w.saveCertFile(base+".ca.crt", data.CaData)
		if err != nil {
			return err
		}
		if crtFile != "" {
			cert.CrtFile = crtFile
		}
		if keyFile != "" {
			cert.KeyFile = keyFile
		}
		if caFile != "" {
			cert.CaFile = caFile
		}
	}
	obj.Cert = cert
	return nil
}

func (w *CeciWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()
	w.out.Info("CeciWorker.Start")

	for _, obj := range w.spec.Proxy {
		if w.isRunning(CeciDir + obj.Id() + ".pid") {
			w.out.Info("CeciWorker.Start: already running %s", obj.Id())
			continue
		}
		w.start(obj)
	}
}

func (w *CeciWorker) Stop(kill bool) {
	w.out.Info("CeciWorker.Stop")
	if !kill {
		return
	}
	for _, obj := range w.spec.Proxy {
		name := CeciDir + obj.Id()
		if w.isRunning(name + ".pid") {
			w.killPid(name + ".pid")
		}
	}
}

func (w *CeciWorker) AddProxy(data schema.CeciProxy) error {
	obj := &co.CeciProxy{
		Mode:    data.Mode,
		Listen:  data.Listen,
		Network: data.Network,
		Target:  data.Target,
	}
	if len(data.Backends) > 0 {
		obj.Backends = make(co.ToForwards, 0, len(data.Backends))
		for _, backend := range data.Backends {
			obj.Backends = append(obj.Backends, &co.ForwardTo{
				Server:   backend.Server,
				Protocol: backend.Protocol,
				Insecure: backend.Insecure,
				Secret:   backend.Secret,
			})
		}
	}
	if err := w.applyCertData(obj, data.Cert); err != nil {
		w.out.Warn("CeciWorker.AddProxy: %s", err)
		return err
	}

	obj.Correct()
	if t, _ := w.spec.FindProxy(obj); t == nil {
		w.spec.AddProxy(obj)
	} else {
		t.Mode = data.Mode
		t.Network = data.Network
		t.Target = data.Target
		t.Backends = obj.Backends
		t.Cert = obj.Cert
	}

	w.restart(obj)
	return nil
}

func (w *CeciWorker) DelProxy(data schema.CeciProxy) {
	obj := &co.CeciProxy{
		Listen: data.Listen,
	}
	obj.Correct()
	if _, removed := w.spec.DelProxy(obj); removed {
		name := CeciDir + obj.Id()
		w.killPid(name + ".pid")
	}
}
