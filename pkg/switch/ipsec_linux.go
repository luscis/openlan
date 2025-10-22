package cswitch

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

const (
	IPSecBin    = "/usr/sbin/ipsec"
	IPSecEtcDir = "/etc/ipsec.d"
	IPSecLogDir = "/var/openlan/ipsec"
)

type IPSecWorker struct {
	*WorkerImpl
	spec *co.IPSecSpecifies
}

func NewIPSecWorker(c *co.Network) *IPSecWorker {
	w := &IPSecWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	w.spec, _ = c.Specifies.(*co.IPSecSpecifies)
	return w
}

var ipsecTmpl = map[string]string{
	"vxlan": `
conn {{ .Name }}
    keyexchange=ike
    ikev2=no
    type=transport
    left={{ .Left }}
{{- if .LeftPort }}
    leftikeport={{ .LeftPort }}
{{- end }}
    right={{ .Right }}
{{- if .RightPort }}
    rightikeport={{ .RightPort }}
{{- end }}
    authby=secret

conn {{ .Name }}-c1
    auto=add
    also={{ .Name }}
{{- if .LeftId }}
    leftid=@c1.{{ .LeftId }}.{{ .Transport }}
{{- end }}
{{- if .RightId }}
    rightid=@c2.{{ .RightId }}.{{ .Transport }}
{{- end }}
    leftprotoport=udp/8472
    rightprotoport=udp

conn {{ .Name }}-c2
    auto=add
    also={{ .Name }}
{{- if .LeftId }}
    leftid=@c2.{{ .LeftId }}.{{ .Transport }}
{{- end }}
{{- if .RightId }}
    rightid=@c1.{{ .RightId }}.{{ .Transport }}
{{- end }}
    leftprotoport=udp
    rightprotoport=udp/8472`,
	"gre": `
conn {{ .Name }}-c1
    auto=add
    ikev2=no
    type=transport
    left={{ .Left }}
{{- if .LeftPort }}
    leftikeport={{ .LeftPort }}
{{- end }}
{{- if .LeftId }}
    leftid=@{{ .LeftId }}.{{ .Transport }}
{{- end }}
    right={{ .Right }}
{{- if .RightId }}
    rightid=@{{ .RightId }}.{{ .Transport }}
{{- end }}
{{- if .RightPort }}
    rightikeport={{ .RightPort }}
{{- end }}
    authby=secret
    leftprotoport=gre
    rightprotoport=gre`,
	"secret": `
%any @{{ .RightId }}.{{ .Transport }} : PSK "{{ .Secret }}"`,
}

func (w *IPSecWorker) Initialize() {
	w.out.Info("IPSecWorker.Initialize")
	if err := os.Mkdir(IPSecLogDir, 0600); err != nil {
		w.out.Warn("IPSecWorker.Initialize %s", err)
	}
}

func (w *IPSecWorker) saveSec(name, tmpl string, data interface{}) error {
	file := fmt.Sprintf("%s/%s", IPSecEtcDir, name)
	out, err := libol.CreateFile(file)
	if err != nil || out == nil {
		return err
	}
	defer out.Close()
	if obj, err := template.New("main").Parse(tmpl); err != nil {
		return err
	} else {
		if err := obj.Execute(out, data); err != nil {
			return err
		}
	}
	return nil
}

func (w *IPSecWorker) startConn(name string) {
	logFile := fmt.Sprintf("%s/%s.log", IPSecLogDir, name)
	logto, err := libol.CreateFile(logFile)
	if err != nil {
		w.out.Warn("IPSecWorker.startConn %s", err)
		return
	}
	libol.Go(func() {
		defer logto.Close()
		cmd := exec.Command(IPSecBin, "auto", "--start", name)
		cmd.Stdout = logto
		cmd.Stderr = logto
		if err := cmd.Run(); err != nil {
			w.out.Warn("IPSecWorker.startConn: %s", err)
			return
		}
		w.out.Info("IPSecWorker.startConn: %v success", name)
	})
}

func (w *IPSecWorker) restartTunnel(tun *co.IPSecTunnel) {
	name := tun.Name
	switch tun.Transport {
	case "vxlan":
		w.startConn(name + "-c1")
		w.startConn(name + "-c2")
	case "gre":
		w.startConn(name + "-c1")
	}
}

func (w *IPSecWorker) addTunnel(tun *co.IPSecTunnel) error {
	connTmpl := ""
	secTmpl := ""

	name := tun.Name
	switch tun.Transport {
	case "vxlan":
		connTmpl = ipsecTmpl["vxlan"]
		secTmpl = ipsecTmpl["secret"]
	case "gre":
		connTmpl = ipsecTmpl["gre"]
		secTmpl = ipsecTmpl["secret"]
	}

	if secTmpl != "" {
		if err := w.saveSec(name+".secrets", secTmpl, tun); err != nil {
			w.out.Error("WorkerImpl.AddTunnel %s", err)
			return err
		}
		libol.Exec(IPSecBin, "auto", "--rereadsecrets")
	}
	if connTmpl != "" {
		if err := w.saveSec(name+".conf", connTmpl, tun); err != nil {
			w.out.Error("WorkerImpl.AddTunnel %s", err)
			return err
		}
		w.restartTunnel(tun)
	}

	return nil
}

func (w *IPSecWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()
	w.out.Info("IPSecWorker.Start")
	for _, tun := range w.spec.Tunnels {
		w.addTunnel(tun)
	}
}

func (w *IPSecWorker) removeTunnel(tun *co.IPSecTunnel) error {
	name := tun.Name
	switch tun.Transport {
	case "vxlan":
		libol.Exec(IPSecBin, "auto", "--delete", "--asynchronous", name+"-c1")
		libol.Exec(IPSecBin, "auto", "--delete", "--asynchronous", name+"-c2")
	case "gre":
		libol.Exec(IPSecBin, "auto", "--delete", "--asynchronous", name+"-c1")
	}
	cfile := fmt.Sprintf("%s/%s.conf", IPSecEtcDir, name)
	sfile := fmt.Sprintf("%s/%s.secrets", IPSecEtcDir, name)

	if err := libol.FileExist(cfile); err == nil {
		if err := os.Remove(cfile); err != nil {
			w.out.Warn("IPSecWorker.RemoveTunnel %s", err)
		}
	}
	if err := libol.FileExist(sfile); err == nil {
		if err := os.Remove(sfile); err != nil {
			w.out.Warn("IPSecWorker.RemoveTunnel %s", err)
		}
	}
	return nil
}

func (w *IPSecWorker) Status() map[string]string {
	status := make(map[string]string)
	out, err := exec.Command(IPSecBin, "status").CombinedOutput()
	if err != nil {
		w.out.Warn("IPSecWorker.Status: %s", err)
		return status
	}
	lines := strings.Split(string(out), "\n")
	start := false
	for _, line := range lines {
		values := strings.SplitN(line, " ", 8)
		if len(values) < 3 {
			continue
		}
		if values[1] == "Total" && values[2] == "IPsec" {
			break
		}
		if values[1] == "Connection" && values[2] == "list:" {
			start = true
			continue
		}
		if !start {
			continue
		}
		if len(values) < 4 {
			continue
		}
		title := strings.Trim(values[1], ":")
		name := strings.Trim(title, "\"")
		state := strings.Trim(values[3], ";")
		if _, ok := status[name]; !ok {
			status[name] = state
		}
	}
	return status
}

func (w *IPSecWorker) Stop() {
	w.out.Info("IPSecWorker.Stop")
	for _, tun := range w.spec.Tunnels {
		w.removeTunnel(tun)
	}
}

func (w *IPSecWorker) Reload(v api.SwitchApi) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *IPSecWorker) AddTunnel(data schema.IPSecTunnel) {
	cfg := &co.IPSecTunnel{
		Left:      data.Left,
		LeftPort:  data.LeftPort,
		LeftId:    data.LeftId,
		Right:     data.Right,
		RightPort: data.RightPort,
		RightId:   data.RightId,
		Secret:    data.Secret,
		Transport: data.Transport,
	}
	cfg.Correct()
	if w.spec.AddTunnel(cfg) {
		w.addTunnel(cfg)
	}
}

func (w *IPSecWorker) DelTunnel(data schema.IPSecTunnel) {
	cfg := &co.IPSecTunnel{
		Left:      data.Left,
		Right:     data.Right,
		Secret:    data.Secret,
		Transport: data.Transport,
	}
	cfg.Correct()
	if _, removed := w.spec.DelTunnel(cfg); removed {
		w.removeTunnel(cfg)
	}
}

func (w *IPSecWorker) StartTunnel(data schema.IPSecTunnel) {
	cfg := &co.IPSecTunnel{
		Left:      data.Left,
		Right:     data.Right,
		Secret:    data.Secret,
		Transport: data.Transport,
	}
	cfg.Correct()
	if _, index := w.spec.FindTunnel(cfg); index != -1 {
		w.restartTunnel(cfg)
	}
}

func (w *IPSecWorker) ListTunnels(call func(obj schema.IPSecTunnel)) {
	status := w.Status()
	for _, tun := range w.spec.Tunnels {
		obj := schema.IPSecTunnel{
			Left:      tun.Left,
			LeftId:    tun.LeftId,
			LeftPort:  tun.LeftPort,
			Right:     tun.Right,
			RightId:   tun.RightId,
			RightPort: tun.RightPort,
			Secret:    tun.Secret,
			Transport: tun.Transport,
		}
		if state, ok := status[tun.Name+"-c1"]; ok {
			obj.State = state
		}
		call(obj)
	}
}
