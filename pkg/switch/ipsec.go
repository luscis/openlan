package cswitch

import (
	"fmt"
	"os"
	"text/template"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
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

// conn vxlan11252
//     keyexchange=ike
//     ikev2=no
//     type=transport
//     left=%defaultroute
//     right=45.135.117.235
//     rightikeport=4501
//     authby=secret

// conn vxlan11252-c2
//     auto=start
//     also=vxlan11252
//     leftid=@c2.vxlan11252.com
//     leftprotoport=udp/8472
//     rightprotoport=udp

// conn vxlan11252-c1
//     auto=start
//     also=vxlan11252
//     leftid=@c1.vxlan11252.com
//     leftprotoport=udp
//     rightprotoport=udp/8472

// conn vxlan11252
//     keyexchange=ike
//     ikev2=no
//     type=transport
//     left=%defaultroute
//     leftikeport=4501
//     right=218.94.118.90
//     authby=secret

// conn vxlan11252-c1
//     auto=start
//     also=vxlan11252
//     rightid=@c1.vxlan11252.com
//     leftprotoport=udp/8472
//     rightprotoport=udp

// conn vxlan11252-c2
//     auto=start
//     also=vxlan11252
//     rightid=@c2.vxlan11252.com
//     leftprotoport=udp
//     rightprotoport=udp/8472

const (
	vxlanTmpl = `
conn {{ .Name }}
    keyexchange=ike
    ikev2=no
    type=transport
    left={{ .Left }}
{{- if .LeftId }}
    leftid={{ .LeftId }}
{{- end }}
{{- if .LeftPort }}
    leftikeport={{ .LeftPort }}
{{- end }}
    right={{ .Right }}
{{- if .RightId }}
    rightid={{ .RightId }}
{{- end }}
{{- if .RightPort }}
    rightikeport={{ .RightPort }}
{{- end }}
    authby=secret

conn {{ .Name }}-c1
    auto=start
    also={{ .Name }}
    leftprotoport=udp/8472
    rightprotoport=udp

conn {{ .Name }}-c2
    auto=start
    also={{ .Remote }}-{{ .Protocol }}
    leftprotoport=udp
    rightprotoport=udp/8472
`
	greTmpl = `
conn {{ .Name }}-c1
    auto=start
    ikev2=insist
    type=transport
    left={{ .Left }}
    right={{ .Right }}
    authby=secret
    leftprotoport=gre
    rightprotoport=gre
`
	secretTmpl = `
%any {{ .Remote }} : PSK "{{ .Secret }}"
`
)

func (w *IPSecWorker) Initialize() {
	w.out.Info("IPSecWorker.Initialize")
}

func (w *IPSecWorker) saveSec(name, tmpl string, data interface{}) error {
	file := fmt.Sprintf("/etc/ipsec.d/%s", name)
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
	promise := libol.NewPromise()
	promise.Go(func() error {
		if out, err := libol.Exec("ipsec", "auto", "--start", "--asynchronous", name); err != nil {
			w.out.Warn("IPSecWorker.startConn: %v %s", out, err)
			return err
		}
		w.out.Info("IPSecWorker.startConn: %v success", name)
		return nil
	})
}

func (w *IPSecWorker) AddTunnel(tunnel *co.IPSecTunnel) error {
	connTmpl := ""
	secTmpl := ""

	name := fmt.Sprintf("%s-%s", tunnel.Right, tunnel.Transport)
	if tunnel.Transport == "vxlan" {
		connTmpl = vxlanTmpl
		secTmpl = secretTmpl
	} else if tunnel.Transport == "gre" {
		connTmpl = greTmpl
		secTmpl = secretTmpl
	}

	if secTmpl != "" {
		if err := w.saveSec(name+".secrets", secTmpl, tunnel); err != nil {
			w.out.Error("WorkerImpl.AddTunnel %s", err)
			return err
		}
		libol.Exec("ipsec", "auto", "--rereadsecrets")
	}
	if connTmpl != "" {
		if err := w.saveSec(name+".conf", connTmpl, tunnel); err != nil {
			w.out.Error("WorkerImpl.AddTunnel %s", err)
			return err
		}
		if tunnel.Transport == "vxlan" {
			w.startConn(name + "-c1")
			w.startConn(name + "-c2")
		} else if tunnel.Transport == "gre" {
			w.startConn(name + "-c1")
		}
	}

	return nil
}

func (w *IPSecWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.out.Info("IPSecWorker.Start")
}

func (w *IPSecWorker) RemoveTunnel(tunnel *co.IPSecTunnel) error {

	name := fmt.Sprintf("%s-%s", tunnel.Right, tunnel.Transport)
	if tunnel.Transport == "vxlan" {
		libol.Exec("ipsec", "auto", "--delete", "--asynchronous", name+"-c1")
		libol.Exec("ipsec", "auto", "--delete", "--asynchronous", name+"-c2")
	} else if tunnel.Transport == "gre" {
		libol.Exec("ipsec", "auto", "--delete", "--asynchronous", name+"-c1")
	}

	cfile := fmt.Sprintf("/etc/ipsec.d/%s.conf", name)
	sfile := fmt.Sprintf("/etc/ipsec.d/%s.secrets", name)

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

func (w *IPSecWorker) Stop() {
	w.out.Info("IPSecWorker.Stop")
}

func (w *IPSecWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}
