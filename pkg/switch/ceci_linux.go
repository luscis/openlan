package cswitch

import (
	"os"
	"os/exec"
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
	runtimeCert, err := w.prepareRuntimeCert(name, obj.Cert)
	if err != nil {
		w.out.Warn("CeciWorker.start: %s", err)
		return
	}

	var data any
	switch obj.Mode {
	case "http":
		data = &co.HttpProxy{
			Listen:   obj.Listen,
			Network:  obj.Network,
			Backends: obj.Backends,
			Cert:     runtimeCert,
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
			w.out.Warn("CeciWorker.start: %s %s", args, err)
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
	if err := os.WriteFile(name, []byte(data+"\n"), 0600); err != nil {
		return "", err
	}
	return name, nil
}

func normalizeCeciCert(data *schema.Cert) (*co.Cert, error) {
	if data == nil {
		return nil, nil
	}

	cert := &co.Cert{
		CrtFile:  data.CrtFile,
		KeyFile:  data.KeyFile,
		CaFile:   data.CaFile,
		CrtData:  data.CrtData,
		KeyData:  data.KeyData,
		CaData:   data.CaData,
		Insecure: data.Insecure,
	}
	if err := cert.LoadData(); err != nil {
		return nil, err
	}
	if cert.CrtData == "" || cert.KeyData == "" {
		return nil, libol.NewErr("certificate and key are required")
	}
	return cert, nil
}

func (w *CeciWorker) prepareRuntimeCert(name string, cert *co.Cert) (*co.Cert, error) {
	if cert == nil {
		return nil, nil
	}
	current := *cert
	if err := current.LoadData(); err != nil {
		return nil, err
	}
	if current.CrtData == "" || current.KeyData == "" {
		return nil, libol.NewErr("certificate content is required")
	}
	crtFile, err := w.saveCertFile(name+".crt", current.CrtData)
	if err != nil {
		return nil, err
	}
	keyFile, err := w.saveCertFile(name+".key", current.KeyData)
	if err != nil {
		return nil, err
	}
	runtime := &co.Cert{
		CrtFile:  crtFile,
		KeyFile:  keyFile,
		Insecure: current.Insecure,
	}
	if current.CaData != "" {
		caFile, err := w.saveCertFile(name+".ca.crt", current.CaData)
		if err != nil {
			return nil, err
		}
		runtime.CaFile = caFile
	}
	return runtime, nil
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
	if cert, err := normalizeCeciCert(data.Cert); err != nil {
		w.out.Warn("CeciWorker.AddProxy: %s", err)
		return err
	} else {
		obj.Cert = cert
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

func (w *CeciWorker) RestartProxy(data schema.CeciProxy) error {
	obj := &co.CeciProxy{
		Listen: data.Listen,
	}
	obj.Correct()
	if current, _ := w.spec.FindProxy(obj); current != nil {
		w.restart(current)
		return nil
	}
	return libol.NewErr("ceci entry not found")
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
