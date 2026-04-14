package cswitch

import (
	"os"
	"os/exec"
	"path/filepath"
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
	cmd := exec.Command(kill, string(pid))
	if err := cmd.Run(); err != nil {
		w.out.Warn("CeciWorker.killPid:%s: %s", pid, err)
		return
	}
}

func (w *CeciWorker) reload(obj *co.CeciProxy) {
	name := CeciDir + obj.Id()

	w.killPid(name + ".pid")
	var data any
	switch obj.Mode {
	case "http":
		data = &co.HttpProxy{
			Listen:   obj.Listen,
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
		w.out.Warn("CeciWorker.reload: %s", err)
		return
	}

	libol.Go(func() {
		out, err := libol.CreateFile(name + ".log")
		if err != nil {
			w.out.Warn("CeciWorker.reload: %s", err)
			return
		}
		defer out.Close()
		w.out.Info("CeciWorker.reload: %s", obj.Id())
		cmd := exec.Command(CeciBin, "-mode", obj.Mode, "-conf", name+".yaml", "-write-pid", name+".pid")
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			w.out.Warn("CeciWorker.reload: %s", err)
			return
		}
	})
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
		w.reload(obj)
	}
}

func (w *CeciWorker) Stop(kill bool) {
	w.out.Info("CeciWorker.Stop")
}

func (w *CeciWorker) AddProxy(data schema.CeciProxy) error {
	obj := &co.CeciProxy{
		Mode:   data.Mode,
		Listen: data.Listen,
		Target: data.Target,
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
		t.Target = data.Target
		t.Backends = obj.Backends
		t.Cert = obj.Cert
	}

	w.reload(obj)
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
