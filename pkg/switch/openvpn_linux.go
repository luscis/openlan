package cswitch

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

const (
	OpenVPNBin = "openvpn"
	VPNCurDir  = "/var/openlan/openvpn/default"
)

type OpenVPNData struct {
	Local                 string
	Port                  string
	CertNot               bool
	Ca                    string
	Cert                  string
	Key                   string
	DhPem                 string
	TlsAuth               string
	Cipher                string
	Server                string
	Device                string
	Protocol              string
	Script                string
	Routes                []string
	Renego                int
	Stats                 string
	IpIp                  string
	Push                  []string
	ClientConfigDir       string
	ClientStatusScriptDir string
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
status {{ .Protocol }}{{ .Port }}server.status 2
client-connect "{{ .ClientStatusScriptDir }}/client-connect.sh"
client-disconnect "{{ .ClientStatusScriptDir }}/client-disconnect.sh"
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
status {{ .Protocol }}{{ .Port }}server.status 2
client-config-dir {{ .ClientConfigDir }}
verb 3
`
	clientConnectScriptTmpl = `#!/bin/bash
log_file="{{ .ClientStatusScriptDir }}/{{ .Protocol }}{{ .Port }}ivplat.status"
if [ -n "$common_name" ]; then
    if grep -q "^$common_name," "$log_file"; then
        sed -i "s/^$common_name,.*/$common_name,$IV_PLAT/" "$log_file"
    else
        if [ -z "$IV_PLAT" ]; then
            IV_PLAT="Unknown"
        fi
        echo "$common_name,$IV_PLAT" >> "$log_file"
    fi
fi
`
	clientDisConnectScriptTmpl = `#!/bin/bash
log_file="{{ .ClientStatusScriptDir }}/{{ .Protocol }}{{ .Port }}ivplat.status"
sed -i "/^$common_name,/d" "$log_file"
`
)

func parseOpenVPNVersion(input string) int {
	re := regexp.MustCompile(`(?i)^openvpn (\d+\.\d+\.\d+)`)
	if match := re.FindStringSubmatch(input); len(match) > 1 {
		version := match[1]
		parts := strings.SplitN(version, ".", 3)
		major, _ := strconv.Atoi(parts[0])
		minor, _ := strconv.Atoi(parts[1])
		return major*10 + minor
	}
	return 0
}

func GetOpenVPNVersion() int {
	out, _ := libol.Exec(OpenVPNBin, "--version")
	return parseOpenVPNVersion(out)
}

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
	if cfg.Version > 24 {
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
	data.ClientStatusScriptDir = obj.ClientIvplatDir()
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
		return VPNCurDir
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
	_ = os.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	return tmplStr
}

func (o *OpenVPN) ClientConnectScriptTmpl() string {
	tmplStr := clientConnectScriptTmpl

	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"connectivplat.tmpl")
	_ = os.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	return tmplStr
}

func (o *OpenVPN) ClientDisConnectScriptTmpl() string {
	tmplStr := clientDisConnectScriptTmpl

	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"disconnectivplat.tmpl")
	_ = os.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
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

func (o *OpenVPN) Pid(full bool) string {
	if data, err := os.ReadFile(o.FilePid(true)); err != nil {
		o.out.Debug("OpenVPN.Stop %s", err)
		return ""
	} else {
		return strings.TrimSpace(string(data))
	}
}

func (o *OpenVPN) DirectoryClientConfig() string {
	if o.Cfg == nil {
		return path.Join(VPNCurDir, "ccd")
	}
	return path.Join(o.Cfg.Directory, "ccd")
}

func (o *OpenVPN) ClientIvplatDir() string {
	if o.Cfg == nil {
		return VPNCurDir
	}
	return o.Cfg.Directory
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
	if data.ClientStatusScriptDir != "" {
		_ = o.writeClientStatusScripts(data)
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
	o.cleanClientConfig()
	for _, fic := range o.Cfg.Clients {
		if fic.Name == "" || fic.Address == "" {
			continue
		}
		ficFile := filepath.Join(ccd, fic.Name)
		pushIP := fmt.Sprintf("ifconfig-push %s %s", fic.Address, fic.Netmask)
		if err := os.WriteFile(ficFile, []byte(pushIP), 0600); err != nil {
			o.out.Warn("OpenVPN.writeClientConfig %s", err)
		}
	}
	return nil
}

func (o *OpenVPN) cleanClientConfig() {
	ccd := o.DirectoryClientConfig()
	files, err := filepath.Glob(path.Join(ccd, "*"))
	if err != nil {
		libol.Warn("OpenVPN.cleanClientConfig %v", err)
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			o.out.Warn("OpenVPN.cleanClientConfig %s", err)
		}
	}
}

func (o *OpenVPN) AddClient(name, address string) error {
	if o.Cfg.AddClient(name, address) {
		o.writeClientConfig()
	}
	return nil
}

func (o *OpenVPN) DelClient(name string) error {
	if _, ok := o.Cfg.DelClient(name); ok {
		o.writeClientConfig()
	}
	return nil
}

func (o *OpenVPN) ListClients(call func(name, address string)) {
	o.Cfg.ListClients(call)
}

func createExecutableFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
}

func (o *OpenVPN) writeClientStatusScripts(data *OpenVPNData) error {
	// make client dir and config file
	cid := o.ClientIvplatDir()
	if err := os.Mkdir(cid, 0600); err != nil {
		o.out.Info("OpenVPN.writeClientStatusScripts %s", err)
	}
	clientConnectScriptFile := filepath.Join(cid, "client-connect.sh")
	fp, err := createExecutableFile(clientConnectScriptFile)
	if err != nil || fp == nil {
		return err
	}
	defer fp.Close()

	tmplStr := o.ClientConnectScriptTmpl()
	if tmpl, err := template.New("clientScript").Parse(tmplStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp, data); err != nil {
			return err
		}
	}

	clientDisConnectFile := filepath.Join(cid, "client-disconnect.sh")
	fp2, err := createExecutableFile(clientDisConnectFile)
	if err != nil || fp2 == nil {
		return err
	}
	defer fp2.Close()

	tmplDisConnectStr := o.ClientDisConnectScriptTmpl()
	if tmpl, err := template.New("clientDisConnectScript").Parse(tmplDisConnectStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp2, data); err != nil {
			return err
		}
	}
	return nil
}

func (o *OpenVPN) Clean() {
	o.cleanClientConfig()
	files := []string{o.FileStats(true), o.FileIpp(true), o.FileClient(true)}
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
	if o.Cfg.Version == 0 {
		o.Cfg.Version = GetOpenVPNVersion()
	}
	o.out.Info("OpenVPN.Initialize version: %d", o.Cfg.Version)

	if err := os.Mkdir(o.Directory(), 0600); err != nil {
		o.out.Info("OpenVPN.Initialize %s", err)
	}
	if err := o.WriteConf(o.FileCfg(true)); err != nil {
		o.out.Warn("OpenVPN.Initialize %s", err)
		return
	}
	if ctx, err := o.Profile(); err == nil {
		file := o.FileClient(true)
		if err := os.WriteFile(file, ctx, 0600); err != nil {
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
	if data, err := os.ReadFile(o.FilePid(true)); err != nil {
		o.out.Debug("OpenVPN.Stop %s", err)
	} else {
		killPath, err := exec.LookPath("kill")
		if err != nil {
			o.out.Warn("kill cmd not found :", err)
		}
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command(killPath, pid)
		if err := cmd.Run(); err != nil {
			o.out.Warn("OpenVPN.Stop %s: %s", pid, err)
		}
	}
	o.Clean()
}

func (o *OpenVPN) checkAlreadyClose(pid string) {
	timeout := 10 * time.Second

	if pid != "" {
		ticker := time.Tick(200 * time.Millisecond)
		timer := time.After(timeout)
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			return
		}
		for {
			select {
			case <-ticker:

				running := libol.IsProcessRunning(pidInt)
				if !running {
					o.out.Debug("OpenVPN.CheckAlreadyClose:vpn is close")
					return
				}
			case <-timer:
				o.out.Warn("OpenVPN.CheckAlreadyClose:vpn close timeout")
				return
			}
		}

	}
}

func (o *OpenVPN) ProfileTmpl() string {
	tmplStr := xAuthClientProfile
	if o.Cfg.Auth == "cert" {
		tmplStr = certClientProfile
	}

	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"client.tmpl")
	_ = os.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
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
	if ctx, err := os.ReadFile(cfg.RootCa); err == nil {
		data.Ca = string(ctx)
	}
	if ctx, err := os.ReadFile(cfg.TlsAuth); err == nil {
		data.TlsAuth = string(ctx)
	}
	return data
}
