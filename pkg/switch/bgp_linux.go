package cswitch

import (
	"os/exec"
	"text/template"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

const (
	BgpBin = "/var/openlan/script/frr-reload"
	BgpEtc = "/etc/frr/frr.conf"
)

type BgpWorker struct {
	*WorkerImpl
	spec *co.BgpSpecifies
}

func NewBgpWorker(c *co.Network) *BgpWorker {
	w := &BgpWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	w.spec, _ = c.Specifies.(*co.BgpSpecifies)
	return w
}

var BgpTmpl = `! GENERATE BY OPENALN
{{- range .Neighbors }}
!
ip prefix-list {{ .Address }}-out seq 10 permit any
ip prefix-list {{ .Address }}-in seq 10 permit any
{{- end }}
!
router bgp {{ .LocalAs }}
 bgp router-id {{ .RouteId }}
 no bgp default ipv4-unicast
 {{- range .Neighbors }}
 neighbor {{ .Address }} remote-as {{ .RemoteAs }}
 {{- end }}
 !
 address-family ipv4 unicast
  redistribute connected
  redistribute kernel
  {{- range .Neighbors }}
  neighbor {{ .Address }} activate
  neighbor {{ .Address }} route-map {{ .Address }}-in in
  neighbor {{ .Address }} route-map {{ .Address }}-out out
  {{- end }}
 exit-address-family
exit
{{- range .Neighbors }}
!
route-map {{ .Address }}-in permit 10
 match ip address prefix-list {{ .Address }}-in
exit
{{- end }}
{{- range .Neighbors }}
!
route-map {{ .Address }}-out permit 10
 match ip address prefix-list {{ .Address }}-out
exit
{{- end }}
!
`

func (w *BgpWorker) Initialize() {
	w.out.Info("BgpWorker.Initialize")
}

func (w *BgpWorker) save() {
	file := BgpEtc
	out, err := libol.CreateFile(file)
	if err != nil || out == nil {
		return
	}
	defer out.Close()
	if obj, err := template.New("main").Parse(BgpTmpl); err != nil {
		w.out.Warn("BgpWorker.save: %s", err)
	} else {
		if err := obj.Execute(out, w.spec); err != nil {
			w.out.Warn("BgpWorker.save: %s", err)
		}
	}
}

func (w *BgpWorker) reload() {
	w.save()
	cmd := exec.Command(BgpBin)
	if err := cmd.Run(); err != nil {
		w.out.Warn("BgpWorker.reload: %s", err)
		return
	}
}

func (w *BgpWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.out.Info("BgpWorker.Start")
	w.reload()
}

func (w *BgpWorker) Stop() {
	w.out.Info("BgpWorker.Stop")
}

func (w *BgpWorker) Enable(data schema.Bgp) {
	w.spec.LocalAs = data.LocalAs
	w.spec.RouteId = data.RouteId
	w.reload()
}

func (w *BgpWorker) Disable() {
}

func (w *BgpWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *BgpWorker) AddNeighbor(data schema.BgpNeighbor) {
	cfg := &co.BgpNeighbor{
		Address:  data.Address,
		RemoteAs: data.RemoteAs,
	}
	cfg.Correct()
	if w.spec.AddNeighbor(cfg) {
		w.reload()
	}
}

func (w *BgpWorker) DelNeighbor(data schema.BgpNeighbor) {
	cfg := &co.BgpNeighbor{
		Address:  data.Address,
		RemoteAs: data.RemoteAs,
	}
	cfg.Correct()
	if _, removed := w.spec.DelNeighbor(cfg); removed {
		w.reload()
	}
}

func (w *BgpWorker) ListNeighbor(call func(obj schema.BgpNeighbor)) {
	for _, nei := range w.spec.Neighbors {
		obj := schema.BgpNeighbor{
			Address:  nei.Address,
			RemoteAs: nei.RemoteAs,
		}
		call(obj)
	}
}
