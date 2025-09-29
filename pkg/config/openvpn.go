package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type OpenVPN struct {
	Network   string           `json:"network" yaml:"network"`
	Url       string           `json:"url,omitempty" yaml:"url,omitempty"`
	Directory string           `json:"directory,omitempty" yaml:"directory,omitempty"`
	Listen    string           `json:"listen" yaml:"listen"`
	Protocol  string           `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Subnet    string           `json:"subnet" yaml:"subnet"`
	Device    string           `json:"device" yaml:"device"`
	Version   int              `json:"version,omitempty" yaml:"version,omitempty"`
	Auth      string           `json:"auth,omitempty" yaml:"auth,omitempty"` // xauth or cert.
	DhPem     string           `json:"dhPem,omitempty" yaml:"dhPem,omitempty"`
	RootCa    string           `json:"rootCa,omitempty" yaml:"rootCa,omitempty"`
	ServerCrt string           `json:"cert,omitempty" yaml:"cert,omitempty"`
	ServerKey string           `json:"key,omitempty" yaml:"key,omitempty"`
	TlsAuth   string           `json:"tlsAuth,omitempty" yaml:"tlsAuth,omitempty"`
	Cipher    string           `json:"cipher,omitempty" yaml:"cipher,omitempty"`
	Routes    []string         `json:"-" yaml:"-"`
	Renego    int              `json:"renego,omitempty" yaml:"renego,omitempty"`
	Script    string           `json:"-" yaml:"-"`
	Push      []string         `json:"push,omitempty" yaml:"push,omitempty"`
	Clients   []*OpenVPNClient `json:"clients,omitempty" yaml:"clients,omitempty"`
}

type OpenVPNClient struct {
	Name    string `json:"name" yaml:"name"`
	Address string `json:"address" yaml:"address"`
	Netmask string `json:"netmask" yaml:"netmask"`
}

var defaultVpn = &OpenVPN{
	Protocol:  "tcp",
	Auth:      "xauth",
	Device:    "tun0",
	RootCa:    VarDir("cert/ca.crt"),
	ServerCrt: VarDir("cert/crt"),
	ServerKey: VarDir("cert/key"),
	DhPem:     VarDir("openvpn/dh.pem"),
	TlsAuth:   VarDir("openvpn/ta.key"),
	Cipher:    "AES-256-CBC",
	Script:    "/usr/bin/openlan",
}

func (o *OpenVPN) AuthBin(obj *OpenVPN) string {
	bin := obj.Script
	bin += " -l " + obj.Url
	bin += " user check"
	bin += " --network " + o.Network
	return bin
}

func (o *OpenVPN) Correct(pool, network string) {
	o.Network = network

	if o.Auth == "" {
		o.Auth = defaultVpn.Auth
	}
	if o.Protocol == "" {
		o.Protocol = defaultVpn.Protocol
	}
	if o.DhPem == "" {
		o.DhPem = defaultVpn.DhPem
	}
	if o.RootCa == "" {
		o.RootCa = defaultVpn.RootCa
	}
	if o.ServerCrt == "" {
		o.ServerCrt = defaultVpn.ServerCrt
	}
	if o.ServerKey == "" {
		o.ServerKey = defaultVpn.ServerKey
	}
	if o.TlsAuth == "" {
		o.TlsAuth = defaultVpn.TlsAuth
	}
	if o.Cipher == "" {
		o.Cipher = defaultVpn.Cipher
	}

	if o.Script == "" {
		o.Script = o.AuthBin(defaultVpn)
	}

	o.Directory = VarDir("openvpn", o.Network)
	if !strings.Contains(o.Listen, ":") {
		o.Listen += ":1194"
	}

	_, port := libol.GetHostPort(o.Listen)
	o.Device = fmt.Sprintf("tun%s", port)

	if o.Subnet == "" {
		value, _ := strconv.Atoi(port)
		o.Subnet = fmt.Sprintf("%s.%d.0/24", pool, value&0xff)
	}

	for _, c := range o.Clients {
		if c.Name == "" || c.Address == "" {
			continue
		}
		if !strings.Contains(c.Name, "@") {
			c.Name = c.Name + "@" + o.Network
		}
	}
}

func (o *OpenVPN) AddRedirectDef1() bool {
	var find = -1

	for i, v := range o.Push {
		if strings.HasPrefix(v, "redirect-gateway") {
			find = i
			break
		}
	}
	if find < 0 {
		o.Push = append(o.Push, "redirect-gateway def1")
	}
	return find < 0
}

func (o *OpenVPN) DelRedirectDef1() bool {
	var find = -1

	for i, v := range o.Push {
		if strings.HasPrefix(v, "redirect-gateway") {
			find = i
			break
		}
	}
	if find > -1 {
		o.Push = append(o.Push[:find], o.Push[find+1:]...)
	}
	return find > -1
}
