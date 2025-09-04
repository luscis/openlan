package cswitch

import (
	"os/exec"
	"strings"
	"text/template"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

const (
	BgpBin = "/var/openlan/script/frr-client"
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
!
service integrated-vtysh-config
{{- if .RouterId }}
router bgp {{ .LocalAs }}
 bgp router-id {{ .RouterId }}
 no bgp default ipv4-unicast
 {{- range .Neighbors }}
 neighbor {{ .Address }} remote-as {{ .RemoteAs }}
 {{-  if .Password }}
 neighbor {{ .Address }} password {{ .Password }}
 {{- end }}
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
!

{{- range $nei := .Neighbors }}
{{- range $seq, $prefix := .Advertis }}
ip prefix-list {{ $nei.Address }}-out seq {{ inc $seq }} permit {{ $prefix }} le 32
{{- end }}
ip prefix-list {{ $nei.Address }}-out seq 65535 deny any
{{- range $seq, $prefix := .Receives }}
ip prefix-list {{ $nei.Address }}-in seq {{ inc $seq }} permit {{ $prefix }} le 32
{{- end }}
ip prefix-list {{ $nei.Address }}-in seq 65535 deny any
{{- end }}
!

{{- range .Neighbors }}
route-map {{ .Address }}-in permit 10
 match ip address prefix-list {{ .Address }}-in
!
{{- end }}

{{- range .Neighbors }}
route-map {{ .Address }}-out permit 10
 match ip address prefix-list {{ .Address }}-out
!
{{- end }}
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

	maps := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	if obj, err := template.New("main").Funcs(maps).Parse(BgpTmpl); err != nil {
		w.out.Warn("BgpWorker.save: %s", err)
	} else {
		if err := obj.Execute(out, w.spec); err != nil {
			w.out.Warn("BgpWorker.save: %s", err)
		}
	}
}

func (w *BgpWorker) reload() {
	w.save()
	cmd := exec.Command(BgpBin, "--reload")
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
	w.spec.RouterId = data.RouterId
	w.reload()
}

func (w *BgpWorker) Disable() {
	w.spec.RouterId = ""
	w.spec.LocalAs = 0
	w.reload()
}

func (w *BgpWorker) Get() *schema.Bgp {
	data := &schema.Bgp{
		LocalAs:  w.spec.LocalAs,
		RouterId: w.spec.RouterId,
	}

	show := map[string]struct {
		State string `json:"state"`
	}{}
	out, err := exec.Command(BgpBin, "--show-neighbors").CombinedOutput()
	if err == nil {
		if err := libol.Unmarshal(&show, out); err != nil {
			w.out.Warn("BgpWorker.Get.Status: %s", err)
		}
	} else {
		w.out.Warn("BgpWorker.Get.Status: %s", err)
	}

	for _, nei := range w.spec.Neighbors {
		obj := schema.BgpNeighbor{
			Address:  nei.Address,
			RemoteAs: nei.RemoteAs,
			Password: nei.Password,
			Receives: nei.Receives,
			Advertis: nei.Advertis,
		}
		if state, ok := show[nei.Address]; ok {
			obj.State = strings.ToLower(state.State)
		}
		data.Neighbors = append(data.Neighbors, obj)
	}
	return data
}

func (w *BgpWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *BgpWorker) AddNeighbor(data schema.BgpNeighbor) {
	obj := &co.BgpNeighbor{
		Address:  data.Address,
		RemoteAs: data.RemoteAs,
		Password: data.Password,
	}
	obj.Correct()
	if nei, _ := w.spec.FindNeighbor(obj); nei == nil {
		w.spec.AddNeighbor(obj)
	} else {
		nei.RemoteAs = data.RemoteAs
		nei.Password = data.Password
	}
	w.reload()
}

func (w *BgpWorker) DelNeighbor(data schema.BgpNeighbor) {
	obj := &co.BgpNeighbor{
		Address:  data.Address,
		RemoteAs: data.RemoteAs,
	}
	obj.Correct()
	if _, removed := w.spec.DelNeighbor(obj); removed {
		w.reload()
	}
}

func (w *BgpWorker) AddReceives(data schema.BgpPrefix) {
	obj := &co.BgpNeighbor{
		Address: data.Neighbor,
	}
	if nei, _ := w.spec.FindNeighbor(obj); nei != nil {
		if nei.AddReceives(data.Prefix) {
			w.reload()
		}
	}
}

func (w *BgpWorker) DelReceives(data schema.BgpPrefix) {
	obj := &co.BgpNeighbor{
		Address: data.Neighbor,
	}
	if nei, _ := w.spec.FindNeighbor(obj); nei != nil {
		if nei.DelReceives(data.Prefix) {
			w.reload()
		}
	}
}

func (w *BgpWorker) AddAdvertis(data schema.BgpPrefix) {
	obj := &co.BgpNeighbor{
		Address: data.Neighbor,
	}
	if nei, _ := w.spec.FindNeighbor(obj); nei != nil {
		if nei.AddAdvertis(data.Prefix) {
			w.reload()
		}
	}
}

func (w *BgpWorker) DelAdvertis(data schema.BgpPrefix) {
	obj := &co.BgpNeighbor{
		Address: data.Neighbor,
	}
	if nei, _ := w.spec.FindNeighbor(obj); nei != nil {
		if nei.DelAdvertis(data.Prefix) {
			w.reload()
		}
	}
}
