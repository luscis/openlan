package _switch

import (
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	L2TPBin = "xl2tpd"
	L2TPDir = "/var/openlan/l2tp"
)

type L2TP struct {
	cfg  *co.L2TP
	out  *libol.SubLogger
	uuid string
}

func NewL2TP(cfg *co.L2TP) *L2TP {
	return &L2TP{
		cfg:  cfg,
		out:  libol.NewSubLogger("l2tp"),
		uuid: "l2tp",
	}
}

func (d *L2TP) Initialize() {
	if err := os.Mkdir(L2TPDir, 0600); err != nil {
		d.out.Error("OpenVPN.Initialize %s", err)
	}
}

func (d *L2TP) Conf() *co.L2TP {
	return d.cfg
}

func (d *L2TP) UUID() string {
	return d.uuid
}

func (d *L2TP) Path() string {
	return L2TPBin
}

func (d *L2TP) PidFile() string {
	return filepath.Join(L2TPDir, d.uuid+".pid")
}

func (d *L2TP) ConfFile() string {
	return filepath.Join(L2TPDir, d.uuid+".conf")
}

func (d *L2TP) OptionsFile() string {
	return filepath.Join(L2TPDir, d.uuid+".options")
}

func (d *L2TP) Tmpl() string {
	return `
[global]
listen-addr = {{ .Listen }}
{{- if .Ipsec }}
ipsec saref = yes
{{- end }}

[lns default]
ip range = {{ .StartAt }}-{{ .EndAt }}
local ip = {{ .Local }}
require chap = yes
refuse pap = yes
require authentication = yes
name = LinuxL2TP
ppp debug = yes
pppoptfile = {{ .Option }}
length bit = yes
`
}

func (d *L2TP) OptionsTmpl() string {
	return `
ipcp-accept-local
ipcp-accept-remote
noccp
# noauth
crtscts
mtu 1410
mru 1410
nodefaultroute
debug
lock
{{- range .Options }}
{{ . }}
{{- end }}
`
}

func (d *L2TP) SaveConf() {
	fp, err := libol.CreateFile(d.ConfFile())
	if err != nil || fp == nil {
		libol.Error("L2TP.SaveConf: %s", err)
		return
	}
	defer fp.Close()

	cfg := d.cfg
	tmpl := d.Tmpl()
	data := struct {
		Listen  string
		Local   string
		StartAt string
		EndAt   string
		Option  string
		Ipsec   bool
	}{
		Listen:  "0.0.0.0",
		Local:   cfg.Address,
		StartAt: cfg.Subnet.Start,
		EndAt:   cfg.Subnet.End,
		Option:  d.OptionsFile(),
	}
	if cfg.IpSec == "enable" {
		data.Ipsec = true
	}
	d.Render(fp, tmpl, data)
}

func (d *L2TP) Render(fp *os.File, tmpl string, data interface{}) {
	if tmpl, err := template.New("main").Parse(tmpl); err == nil {
		if err := tmpl.Execute(fp, data); err != nil {
			d.out.Warn("L2TP.Render: %s", err)
			return
		}
	} else {
		d.out.Warn("L2TP.Render: %s", err)
	}
}

func (d *L2TP) SaveOptions() {
	fp, err := libol.CreateFile(d.OptionsFile())
	if err != nil || fp == nil {
		libol.Error("L2TP.SaveOptions: %s", err)
		return
	}
	defer fp.Close()

	cfg := d.cfg
	tmpl := d.OptionsTmpl()
	data := struct {
		Options []string
	}{
		Options: cfg.Options,
	}
	d.Render(fp, tmpl, data)
}

func (d *L2TP) Start() {
	if d.cfg.Subnet == nil {
		return
	}
	d.SaveConf()
	d.SaveOptions()
	libol.Go(func() {
		args := []string{
			"-c", d.ConfFile(),
			"-p", d.PidFile(),
			"-D",
		}
		d.out.Info("L2TP.Start %s %v", d.Path(), args)
		cmd := exec.Command(d.Path(), args...)
		if err := cmd.Run(); err != nil {
			d.out.Error("L2TP.Start %s: %s", d.uuid, err)
		}
	})
}

func (d *L2TP) Clean() {
	files := []string{
		d.PidFile(), d.ConfFile(),
	}
	for _, file := range files {
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				d.out.Warn("L2TP.Clean %s", err)
			}
		}
	}
}

func (d *L2TP) Stop() {
	if d.cfg.Subnet == nil {
		return
	}
	if data, err := ioutil.ReadFile(d.PidFile()); err != nil {
		d.out.Info("L2TP.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			d.out.Warn("L2TP.Stop %s: %s", pid, err)
		}
	}
	d.Clean()
}
