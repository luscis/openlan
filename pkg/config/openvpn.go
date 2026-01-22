package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type OpenVPN struct {
	Network   string           `json:"-" yaml:"-"`
	Url       string           `json:"-" yaml:"-"`
	Directory string           `json:"-" yaml:"-"`
	Listen    string           `json:"listen" yaml:"listen"`
	Protocol  string           `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Subnet    string           `json:"subnet" yaml:"subnet"`
	IPAddr    *net.IPNet       `json:"-" yaml:"-"`
	Device    string           `json:"device" yaml:"device"`
	Version   int              `json:"-" yaml:"-"`
	DhPem     string           `json:"-" yaml:"-"`
	RootCa    string           `json:"-" yaml:"-"`
	ServerCrt string           `json:"-" yaml:"-"`
	ServerKey string           `json:"-" yaml:"-"`
	TlsAuth   string           `json:"-" yaml:"-"`
	Routes    []string         `json:"-" yaml:"-"`
	Renego    int              `json:"-" yaml:"-"`
	Script    string           `json:"-" yaml:"-"`
	Push      []string         `json:"push,omitempty" yaml:"push,omitempty"`
	Clients   []*OpenVPNClient `json:"clients,omitempty" yaml:"clients,omitempty"`
	Cipher    string           `json:"cipher,omitempty" yaml:"cipher,omitempty"` // AES-256-GCM:AES-128-GCM:SM4-GCM:SM4-CBC
}

type OpenVPNClient struct {
	Name    string `json:"name" yaml:"name"`
	Address string `json:"address" yaml:"address"`
	Netmask string `json:"-" yaml:"-"`
}

func (c *OpenVPNClient) Correct(v *OpenVPN) {
	c.Netmask = v.GetNetmask()
	if !strings.Contains(c.Name, "@") {
		c.Name = c.Name + "@" + v.Network
	}
}

var defaultVpn = &OpenVPN{
	Protocol:  "tcp",
	Device:    "tun0",
	RootCa:    VarDir("cert/ca.crt"),
	ServerCrt: VarDir("cert/crt"),
	ServerKey: VarDir("cert/key"),
	DhPem:     VarDir("openvpn/dh.pem"),
	TlsAuth:   VarDir("openvpn/ta.key"),
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
	o.Script = o.AuthBin(defaultVpn)
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
		c.Correct(o)
	}
}

func (o *OpenVPN) GetNetmask() string {
	if o.IPAddr == nil {
		if _, n, err := net.ParseCIDR(o.Subnet); err == nil {
			o.IPAddr = n
		}
	}
	return net.IP(o.IPAddr.Mask).String()
}

func (o *OpenVPN) GetIPAddr() string {
	if o.IPAddr == nil {
		if _, n, err := net.ParseCIDR(o.Subnet); err == nil {
			o.IPAddr = n
		}
	}
	return net.IP(o.IPAddr.IP).String()
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

func (o *OpenVPN) FindClient(name string) (*OpenVPNClient, int) {
	for i, obj := range o.Clients {
		if name == obj.Name {
			return obj, i
		}
	}
	return nil, -1
}

func (o *OpenVPN) AddClient(name, address string) bool {
	value := &OpenVPNClient{
		Name:    name,
		Address: address,
	}
	value.Correct(o)

	_, index := o.FindClient(name)
	if index == -1 {
		o.Clients = append(o.Clients, value)
	}

	return index == -1
}

func (o *OpenVPN) DelClient(name string) (*OpenVPNClient, bool) {
	value := &OpenVPNClient{
		Name: name,
	}
	value.Correct(o)

	obj, index := o.FindClient(value.Name)
	if index != -1 {
		o.Clients = append(o.Clients[:index], o.Clients[index+1:]...)
	}

	return obj, index != -1
}

func (o *OpenVPN) ListClients(call func(name, address string)) {
	for _, obj := range o.Clients {
		call(obj.Name, obj.Address)
	}
}
