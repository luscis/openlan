package cswitch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

const (
	OpenVPNBin    = "openvpn"
	DefaultCurDir = "/var/openlan/openvpn/default"
)

type OpenVPNData struct {
	Local           string
	Port            string
	CertNot         bool
	Ca              string
	Cert            string
	Key             string
	DhPem           string
	TlsAuth         string
	Cipher          string
	Server          string
	Device          string
	Protocol        string
	Script          string
	Routes          []string
	Renego          int
	Stats           string
	IpIp            string
	Push            []string
	ClientConfigDir string
}

const (
	xAuthConfTmpl = `# Generate by OpenLAN
local {{ .Local }}
port {{ .Port }}
proto {{ .Protocol }}
dev {{ .Device }}
reneg-sec {{ .Renego }}
keepalive 10 120
persist-key
persist-tun
ca {{ .Ca }}
cert {{ .Cert }}
key {{ .Key }}
dh {{ .DhPem }}
server {{ .Server }}
{{- range .Routes }}
push "route {{ . }}"
{{- end }}
{{- range .Push }}
push "{{ . }}"
{{- end }}
ifconfig-pool-persist {{ .Protocol }}{{ .Port }}ipp
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
status {{ .Protocol }}{{ .Port }}server.status 5
{{- if .CertNot }}
client-cert-not-required
{{- else }}
verify-client-cert none
{{- end }}
script-security 3
auth-user-pass-verify "{{ .Script }}" via-env
username-as-common-name
client-config-dir {{ .ClientConfigDir }}
verb 3
`
	certConfTmpl = `# Generate by OpenLAN
local {{ .Local }}
port {{ .Port }}
proto {{ .Protocol }}
dev {{ .Device }}
reneg-sec {{ .Renego }}
keepalive 10 120
persist-key
persist-tun
ca {{ .Ca }}
cert {{ .Cert }}
key {{ .Key }}
dh {{ .DhPem }}
server {{ .Server }}
{{- range .Routes }}
push "route {{ . }}"
{{- end }}
ifconfig-pool-persist {{ .Protocol }}{{ .Port }}ipp
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
status {{ .Protocol }}{{ .Port }}server.status 5
client-config-dir {{ .ClientConfigDir }}
verb 3
`
)

func NewOpenVPNDataFromConf(obj *OpenVPN) *OpenVPNData {
	cfg := obj.Cfg
	data := &OpenVPNData{
		Local:    obj.Local,
		Port:     obj.Port,
		CertNot:  true,
		Ca:       cfg.RootCa,
		Cert:     cfg.ServerCrt,
		Key:      cfg.ServerKey,
		DhPem:    cfg.DhPem,
		TlsAuth:  cfg.TlsAuth,
		Cipher:   cfg.Cipher,
		Device:   cfg.Device,
		Protocol: cfg.Protocol,
		Script:   cfg.Script,
		Renego:   cfg.Renego,
		Push:     cfg.Push,
	}
	if cfg.Version > 23 {
		data.CertNot = false
	}
	addr, _ := libol.IPNetwork(cfg.Subnet)
	data.Server = strings.ReplaceAll(addr, "/", " ")
	for _, rt := range cfg.Routes {
		if addr, err := libol.IPNetwork(rt); err == nil {
			r := strings.ReplaceAll(addr, "/", " ")
			data.Routes = append(data.Routes, r)
		}
	}
	data.ClientConfigDir = obj.DirectoryClientConfig()
	return data
}

type OpenVPN struct {
	Cfg      *co.OpenVPN
	out      *libol.SubLogger
	Protocol string
	Local    string
	Port     string
}

func NewOpenVPN(cfg *co.OpenVPN) *OpenVPN {
	obj := &OpenVPN{
		Cfg:      cfg,
		out:      libol.NewSubLogger(cfg.Network),
		Protocol: cfg.Protocol,
		Local:    "0.0.0.0",
		Port:     "4494",
	}
	obj.Local = strings.SplitN(cfg.Listen, ":", 2)[0]
	if strings.Contains(cfg.Listen, ":") {
		obj.Port = strings.SplitN(cfg.Listen, ":", 2)[1]
	}
	return obj
}

func (o *OpenVPN) ID() string {
	return o.Protocol + o.Port
}

func (o *OpenVPN) Path() string {
	return OpenVPNBin
}

func (o *OpenVPN) Directory() string {
	if o.Cfg == nil {
		return DefaultCurDir
	}
	return o.Cfg.Directory
}

func (o *OpenVPN) FileCfg(full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name := o.ID() + "server.conf"
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) FileClient(full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name := o.ID() + "client.ovpn"
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) FileLog(full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name := o.ID() + "server.log"
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) FilePid(full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name := o.ID() + "server.pid"
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) FileStats(full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name := o.ID() + "server.stats"
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) ServerTmpl() string {
	tmplStr := xAuthConfTmpl
	if o.Cfg.Auth == "cert" {
		tmplStr = certConfTmpl
	}
	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"server.tmpl")
	_ = ioutil.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	return tmplStr
}

func (o *OpenVPN) FileIpp(full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name := o.ID() + "ipp"
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) DirectoryClientConfig() string {
	if o.Cfg == nil {
		return path.Join(DefaultCurDir, "ccd")
	}
	return path.Join(o.Cfg.Directory, "ccd")
}

func (o *OpenVPN) WriteConf(path string) error {
	fp, err := libol.CreateFile(path)
	if err != nil || fp == nil {
		return err
	}
	defer fp.Close()
	data := NewOpenVPNDataFromConf(o)
	o.out.Debug("OpenVPN.WriteConf %v", data)
	if data.ClientConfigDir != "" {
		_ = o.writeClientConfig()
	}
	tmplStr := o.ServerTmpl()
	if tmpl, err := template.New("main").Parse(tmplStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp, data); err != nil {
			return err
		}
	}
	return nil
}

func (o *OpenVPN) writeClientConfig() error {
	// make client dir and config file
	ccd := o.DirectoryClientConfig()
	if err := os.Mkdir(ccd, 0600); err != nil {
		o.out.Info("OpenVPN.writeClientConfig %s", err)
	}
	for _, fic := range o.Cfg.Clients {
		if fic.Name == "" || fic.Address == "" {
			continue
		}
		ficFile := filepath.Join(ccd, fic.Name)
		pushIP := fmt.Sprintf("ifconfig-push %s %s", fic.Address, fic.Netmask)
		if err := ioutil.WriteFile(ficFile, []byte(pushIP), 0600); err != nil {
			o.out.Warn("OpenVPN.writeClientConfig %s", err)
		}
	}

	return nil
}

func (o *OpenVPN) Clean() {
	ccd := o.DirectoryClientConfig()
	for _, fic := range o.Cfg.Clients {
		if fic.Name == "" || fic.Address == "" {
			continue
		}
		file := filepath.Join(ccd, fic.Name)
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				o.out.Warn("OpenVPN.Clean %s", err)
			}
		}
	}
	files := []string{o.FileStats(true), o.FileIpp(true)}
	for _, file := range files {
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				o.out.Warn("OpenVPN.Clean %s", err)
			}
		}
	}
}

func (o *OpenVPN) Initialize() {
	if !o.ValidConf() {
		return
	}
	o.Clean()
	if err := os.Mkdir(o.Directory(), 0600); err != nil {
		o.out.Info("OpenVPN.Initialize %s", err)
	}
	if err := o.WriteConf(o.FileCfg(true)); err != nil {
		o.out.Warn("OpenVPN.Initialize %s", err)
		return
	}
	if ctx, err := o.Profile(); err == nil {
		file := o.FileClient(true)
		if err := ioutil.WriteFile(file, ctx, 0600); err != nil {
			o.out.Warn("OpenVPN.Initialize %s", err)
		}
	} else {
		o.out.Warn("OpenVPN.Initialize %s", err)
	}
}

func (o *OpenVPN) ValidConf() bool {
	if o.Cfg == nil {
		return false
	}
	if o.Cfg.Listen == "" || o.Cfg.Subnet == "" {
		return false
	}
	return true
}

func (o *OpenVPN) Start() {
	if !o.ValidConf() {
		return
	}
	log, err := libol.CreateFile(o.FileLog(true))
	if err != nil {
		o.out.Warn("OpenVPN.Start %s", err)
		return
	}
	libol.Go(func() {
		defer log.Close()
		args := []string{
			"--cd", o.Directory(),
			"--config", o.FileCfg(false),
			"--writepid", o.FilePid(false),
		}
		cmd := exec.Command(o.Path(), args...)
		cmd.Stdout = log
		cmd.Stderr = log
		if err := cmd.Run(); err != nil {
			o.out.Error("OpenVPN.Start %s: %s", o.ID(), err)
		}
	})
}

func (o *OpenVPN) Stop() {
	if !o.ValidConf() {
		return
	}
	if data, err := ioutil.ReadFile(o.FilePid(true)); err != nil {
		o.out.Debug("OpenVPN.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			o.out.Warn("OpenVPN.Stop %s: %s", pid, err)
		}
	}
	o.Clean()
}

func (o *OpenVPN) ProfileTmpl() string {
	tmplStr := xAuthClientProfile
	if o.Cfg.Auth == "cert" {
		tmplStr = certClientProfile
	}
	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"client.tmpl")
	_ = ioutil.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	return tmplStr
}

func (o *OpenVPN) Profile() ([]byte, error) {
	data := NewOpenVPNProfileFromConf(o)
	tmplStr := o.ProfileTmpl()
	tmpl, err := template.New("main").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err == nil {
		return out.Bytes(), nil
	} else {
		return nil, err
	}
}

type OpenVPNProfile struct {
	Server   string
	Port     string
	Ca       string
	Cert     string
	Key      string
	TlsAuth  string
	Cipher   string
	Device   string
	Protocol string
	Renego   int
}

const (
	xAuthClientProfile = `# Generate by OpenLAN
client
dev {{ .Device }}
route-metric 300
proto {{ .Protocol }}
remote {{ .Server }} {{ .Port }}
reneg-sec {{ .Renego }}
resolv-retry infinite
nobind
persist-key
persist-tun
<ca>
{{ .Ca -}}
</ca>
remote-cert-tls server
<tls-auth>
{{ .TlsAuth -}}
</tls-auth>
key-direction 1
cipher {{ .Cipher }}
auth-nocache
verb 4
auth-user-pass
`
	certClientProfile = `# Generate by OpenLAN
client
dev {{ .Device }}
route-metric 300
proto {{ .Protocol }}
remote {{ .Server }} {{ .Port }}
reneg-sec {{ .Renego }}
resolv-retry infinite
nobind
persist-key
persist-tun
<ca>
{{ .Ca -}}
</ca>
remote-cert-tls server
<tls-auth>
{{ .TlsAuth -}}
</tls-auth>
key-direction 1
cipher {{ .Cipher }}
auth-nocache
verb 4
`
)

func NewOpenVPNProfileFromConf(obj *OpenVPN) *OpenVPNProfile {
	cfg := obj.Cfg
	data := &OpenVPNProfile{
		Server:   obj.Local,
		Port:     obj.Port,
		Cipher:   cfg.Cipher,
		Device:   cfg.Device[:3],
		Protocol: cfg.Protocol,
		Renego:   cfg.Renego,
	}
	if data.Server == "0.0.0.0" {
		if name, err := os.Hostname(); err == nil {
			data.Server = name
		}
	}
	if ctx, err := ioutil.ReadFile(cfg.RootCa); err == nil {
		data.Ca = string(ctx)
	}
	if ctx, err := ioutil.ReadFile(cfg.TlsAuth); err == nil {
		data.TlsAuth = string(ctx)
	}
	return data
}
