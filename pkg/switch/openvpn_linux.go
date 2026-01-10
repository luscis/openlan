package cswitch

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
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
	Local     string
	Port      string
	CertNot   bool
	Ca        string
	Cert      string
	Key       string
	DhPem     string
	TlsAuth   string
	Cipher    string
	Server    string
	Device    string
	Protocol  string
	Script    string
	Routes    []string
	Renego    int
	Stats     string
	IpIp      string
	Push      []string
	ClientDir string
	ServerDir string
}

const (
	vConfTmpl = `# Generate by OpenLAN
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
management {{ .ServerDir }}/{{ .Protocol }}{{ .Port }}server.sock unix
status {{ .Protocol }}{{ .Port }}server.client 2
client-connect "{{ .ServerDir }}/client-up.sh"
client-disconnect "{{ .ServerDir }}/client-down.sh"
{{- if .CertNot }}
client-cert-not-required
{{- else }}
verify-client-cert none
{{- end }}
script-security 3
auth-user-pass-verify "{{ .Script }}" via-env
username-as-common-name
client-config-dir {{ .ClientDir }}
verb 3
`
	vClientUpTmpl = `#!/bin/bash
log_file="{{ .ServerDir }}/{{ .Protocol }}{{ .Port }}server.plat"
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
	vClientDownTmpl = `#!/bin/bash
log_file="{{ .ServerDir }}/{{ .Protocol }}{{ .Port }}server.plat"
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
	data.ClientDir = obj.ClientDir()
	data.ServerDir = obj.ServerDir()
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

func (o *OpenVPN) tofile(name string, full bool) string {
	if o.Cfg == nil {
		return ""
	}
	name = o.ID() + name
	if !full {
		return name
	}
	return filepath.Join(o.Cfg.Directory, name)
}

func (o *OpenVPN) FileCfg(full bool) string {
	return o.tofile("server.conf", full)
}

func (o *OpenVPN) FileClientProfile(full bool) string {
	return o.tofile("client.ovpn", full)

}

func (o *OpenVPN) FileLog(full bool) string {
	return o.tofile("server.log", full)
}

func (o *OpenVPN) FilePid(full bool) string {
	return o.tofile("server.pid", full)
}

func (o *OpenVPN) FileIpp(full bool) string {
	return o.tofile("ipp", full)
}

func (o *OpenVPN) FileClient(full bool) string {
	return o.tofile("server.client", full)
}

func (o *OpenVPN) FileCtrl(full bool) string {
	return o.tofile("server.sock", full)
}

func (o *OpenVPN) ClientDir() string {
	if o.Cfg == nil {
		return path.Join(VPNCurDir, "ccd")
	}
	return path.Join(o.Cfg.Directory, "ccd")
}

func (o *OpenVPN) ServerDir() string {
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
	o.out.Debug("OpenVPN.WriteConf: %v", data)
	if data.ClientDir != "" {
		_ = o.writeClientConf()
	}
	if data.ServerDir != "" {
		_ = o.writeClientPlat(data)
	}
	tmplStr := vConfTmpl
	if tmpl, err := template.New("main").Parse(tmplStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp, data); err != nil {
			return err
		}
	}
	return nil
}

func (o *OpenVPN) writeClientConf() error {
	// make client dir and config file
	ccd := o.ClientDir()
	if err := os.Mkdir(ccd, 0600); err != nil {
		o.out.Info("OpenVPN.writeClientConf: %s", err)
	}
	o.cleanClientConf()
	for _, fic := range o.Cfg.Clients {
		if fic.Name == "" || fic.Address == "" {
			continue
		}
		ficFile := filepath.Join(ccd, fic.Name)
		pushIP := fmt.Sprintf("ifconfig-push %s %s", fic.Address, fic.Netmask)
		if err := os.WriteFile(ficFile, []byte(pushIP), 0600); err != nil {
			o.out.Warn("OpenVPN.writeClientConf: %s", err)
		}
	}
	return nil
}

func (o *OpenVPN) cleanClientConf() {
	ccd := o.ClientDir()
	files, err := filepath.Glob(path.Join(ccd, "*"))
	if err != nil {
		libol.Warn("OpenVPN.cleanClientConf: %v", err)
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			o.out.Warn("OpenVPN.cleanClientConf: %s", err)
		}
	}
}

func (o *OpenVPN) AddClient(name, address string) error {
	if o.Cfg.AddClient(name, address) {
		o.writeClientConf()
	}
	return nil
}

func (o *OpenVPN) DelClient(name string) error {
	if _, ok := o.Cfg.DelClient(name); ok {
		o.writeClientConf()
	}
	return nil
}

func (o *OpenVPN) ListClients(call func(name, address string)) {
	o.Cfg.ListClients(call)
}

func (o *OpenVPN) writeClientPlat(data *OpenVPNData) error {
	// make client dir and config file
	cid := o.ServerDir()
	if err := os.Mkdir(cid, 0600); err != nil {
		o.out.Info("OpenVPN.writeClientPlat: %s", err)
	}
	clientConnectScriptFile := filepath.Join(cid, "client-up.sh")
	fp, err := libol.CreateFileEx(clientConnectScriptFile)
	if err != nil || fp == nil {
		return err
	}
	defer fp.Close()

	tmplStr := vClientUpTmpl
	if tmpl, err := template.New("clientScript").Parse(tmplStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp, data); err != nil {
			return err
		}
	}

	clientDisConnectFile := filepath.Join(cid, "client-down.sh")
	fp2, err := libol.CreateFileEx(clientDisConnectFile)
	if err != nil || fp2 == nil {
		return err
	}
	defer fp2.Close()

	tmplDisConnectStr := vClientDownTmpl
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
	o.cleanClientConf()
	files := []string{o.FileIpp(true), o.FileClientProfile(true)}
	for _, file := range files {
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				o.out.Warn("OpenVPN.Clean: %s", err)
			}
		}
	}
}

func (o *OpenVPN) Initialize() {
	if !o.ValidConf() {
		return
	}

	if o.FindPid() == 0 {
		o.Clean()
	}
	if o.Cfg.Version == 0 {
		o.Cfg.Version = GetOpenVPNVersion()
	}
	o.out.Info("OpenVPN.Initialize version: %d", o.Cfg.Version)

	if err := os.Mkdir(o.Directory(), 0600); err != nil {
		o.out.Info("OpenVPN.Initialize: %s", err)
	}
	if err := o.WriteConf(o.FileCfg(true)); err != nil {
		o.out.Warn("OpenVPN.Initialize: %s", err)
		return
	}
	if ctx, err := o.Profile(); err == nil {
		file := o.FileClientProfile(true)
		if err := os.WriteFile(file, ctx, 0600); err != nil {
			o.out.Warn("OpenVPN.Initialize: %s", err)
		}
	} else {
		o.out.Warn("OpenVPN.Initialize: %s", err)
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

func (o *OpenVPN) FindPid() int {
	pid := 0
	if v, err := os.ReadFile(o.FilePid(true)); err == nil {
		fmt.Sscanf(string(v), "%d", &pid)
	}
	return pid
}

func (o *OpenVPN) Start() {
	if !o.ValidConf() {
		return
	}

	pid := o.FindPid()
	o.out.Info("OpenVPN.Start: older pid:%d", pid)
	if pid > 0 {
		if ok := libol.HasProcess(pid); ok {
			o.out.Info("OpenVPN.Start: already running")
			return
		}
	}

	libol.Go(func() {
		args := []string{
			"--cd", o.Directory(),
			"--config", o.FileCfg(false),
			"--writepid", o.FilePid(false),
			"--log-append", o.FileLog(false),
		}
		o.out.Info("%s with %s", o.Path(), args)
		cmd := exec.Command(o.Path(), args...)
		if err := cmd.Start(); err != nil {
			o.out.Error("OpenVPN.Start: %s: %s", o.ID(), err)
		}
		cmd.Wait()
	})
}

func (o *OpenVPN) Kill() {
	pid := o.FindPid()
	if pid == 0 {
		return
	}
	o.out.Info("OpenVPN.Kill %d", pid)
	if proc, err := libol.Kill(pid); err == nil {
		proc.Wait()
		if err := os.Remove(o.FilePid(true)); err != nil {
			o.out.Warn("OpenVPN.Kill: %s", err)
		}
	} else {
		o.out.Warn("OpenVPN.Kill: %d %s", pid, err)
	}
}

func (o *OpenVPN) Stop() {
	if !o.ValidConf() {
		return
	}
	if pid := o.FindPid(); libol.HasProcess(pid) {
		o.out.Info("OpenVPN.Stop: without kill %d.", pid)
	} else {
		o.Clean()
	}
}

func (o *OpenVPN) CheckWait() {
	timeout := 10 * time.Second
	if pid := o.FindPid(); pid > 0 {
		ticker := time.Tick(200 * time.Millisecond)
		timer := time.After(timeout)
		for {
			select {
			case <-ticker:
				running := libol.HasProcess(pid)
				if !running {
					o.out.Debug("OpenVPN.checkWait: vpn is close")
					return
				}
			case <-timer:
				o.out.Warn("OpenVPN.checkWait: vpn close timeout")
				return
			}
		}
	}
}

func (o *OpenVPN) Profile() ([]byte, error) {
	data := NewOpenVPNProfileFromConf(o)
	tmplStr := vClientProfile
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

func (o *OpenVPN) Exec(cmd string) error {
	conn, err := net.Dial("unix", o.FileCtrl(true))
	if err != nil {
		libol.Warn("OpenVPN.Exec: %v", err)
		return err
	}
	defer conn.Close()

	// Read help info.
	r := bufio.NewReader(conn)
	out, _, err := r.ReadLine()
	if err != nil {
		libol.Warn("OpenVPN.Exec: %v", err)
		return err
	}

	libol.Info("OpenVPN.Exec: %s", cmd)
	_, err = fmt.Fprintln(conn, cmd)
	if err != nil {
		libol.Warn("OpenVPN.Exec: %v", err)
		return err
	}
	out, _, err = r.ReadLine()
	if err != nil {
		libol.Warn("OpenVPN.Exec: %v", err)
		return err
	}
	libol.Info("OpenVPN.Exec: %s", out)

	return nil
}

func (o *OpenVPN) KillClient(name string) error {
	cmd := fmt.Sprintf("kill %s", name)
	return o.Exec(cmd)
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
	vClientProfile = `# Generate by OpenLAN
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
auth-nocache
verb 4
auth-user-pass
`
)

func NewOpenVPNProfileFromConf(obj *OpenVPN) *OpenVPNProfile {
	cfg := obj.Cfg
	data := &OpenVPNProfile{
		Server:   obj.Local,
		Port:     obj.Port,
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
