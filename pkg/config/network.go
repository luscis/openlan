package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type DNAT struct {
	Protocol string `json:"protocol" yaml:"protocol"`
	Dest     string `json:"destination,omitempty" yaml:"destination,omitempty"`
	Dport    int    `json:"dport" yaml:"dport"`
	ToDest   string `json:"todestination" yaml:"todestination"`
	ToDport  int    `json:"todport" yaml:"todport"`
}

func (d *DNAT) Id() string {
	return fmt.Sprintf("%s:%s:%d", d.Protocol, d.Dest, d.Dport)
}

func (d *DNAT) Correct() {
	if d.ToDport == 0 {
		d.ToDport = d.Dport
	}
}

type Network struct {
	ConfDir   string              `json:"-" yaml:"-"`
	File      string              `json:"-" yaml:"-"`
	Alias     string              `json:"-" yaml:"-"`
	Name      string              `json:"name" yaml:"name"`
	Provider  string              `json:"provider,omitempty" yaml:"provider,omitempty"`
	Bridge    *Bridge             `json:"bridge,omitempty" yaml:"bridge,omitempty"`
	Subnet    *Subnet             `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	OpenVPN   *OpenVPN            `json:"openvpn,omitempty" yaml:"openvpn,omitempty"`
	Links     []Access            `json:"links,omitempty" yaml:"links,omitempty"`
	Hosts     []HostLease         `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Routes    []PrefixRoute       `json:"routes,omitempty" yaml:"routes,omitempty"`
	Acl       string              `json:"acl,omitempty" yaml:"acl,omitempty"`
	Specifies interface{}         `json:"specifies,omitempty" yaml:"specifies,omitempty"`
	Dhcp      string              `json:"dhcp,omitempty" yaml:"dhcp,omitempty"`
	Outputs   []*Output           `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	ZTrust    string              `json:"ztrust,omitempty" yaml:"ztrust,omitempty"`
	Qos       string              `json:"qos,omitempty" yaml:"qos,omitempty"`
	Snat      string              `json:"snat,omitempty" yaml:"snat,omitempty"`
	Namespace string              `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	FindHop   map[string]*FindHop `json:"findhop,omitempty" yaml:"findhop,omitempty"`
	Dnat      []*DNAT             `json:"dnat,omitempty" yaml:"dnat,omitempty"`
	AddrPool  string              `json:"-" yaml:"-"`
}

func (n *Network) NewSpecifies() any {
	switch n.Provider {
	case "ipsec":
		n.Specifies = &IPSecSpecifies{}
	case "router":
		n.Specifies = &RouterSpecifies{}
	case "bgp":
		n.Specifies = &BgpSpecifies{}
	case "ceci":

		n.Specifies = &CeciSpecifies{}
	default:
		n.Specifies = nil
	}
	if n.Specifies != nil {
		n.Name = n.Provider
	}
	return n.Specifies
}

func (n *Network) Init() {
	switch n.Provider {
	case "router", "ipsec", "bgp", "ceci":
		// Special network.
	default:
		if n.Bridge == nil {
			n.Bridge = &Bridge{}
		}
		if n.Subnet == nil {
			n.Subnet = &Subnet{}
		}
		n.Bridge.Network = n.Name
	}
}

func (n *Network) Correct() {
	ipAddr := ""
	ipMask := ""

	n.Init()
	if n.Snat == "" {
		n.Snat = "enable"
	}

	switch n.Provider {
	case "router":
		spec := n.Specifies
		if obj, ok := spec.(*RouterSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	case "ipsec":
		spec := n.Specifies
		if obj, ok := spec.(*IPSecSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	case "bgp":
		spec := n.Specifies
		if obj, ok := spec.(*BgpSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	case "ceci":
		spec := n.Specifies
		if obj, ok := spec.(*CeciSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	default:
		n.Bridge.Correct()
		if _i, _n, err := net.ParseCIDR(n.Bridge.Address); err == nil {
			ipAddr = _i.String()
			ipMask = net.IP(_n.Mask).String()
		}
		if n.Subnet.Netmask == "" {
			n.Subnet.Netmask = ipMask
		}
	}

	CorrectRoutes(n.Routes, ipAddr)
	if n.OpenVPN != nil {
		n.OpenVPN.Correct(n.AddrPool, n.Name)
	}

	for key, value := range n.FindHop {
		value.Correct()
		n.FindHop[key] = value
	}
	for _, value := range n.Outputs {
		value.Correct()
	}
}

func (n *Network) Dir(module string) string {
	var file string

	if n.IsYaml() {
		file = n.Name + ".yaml"
	} else {
		file = n.Name + ".json"
	}

	return filepath.Join(n.ConfDir, module, file)
}

func (n *Network) IsYaml() bool {
	return libol.IsYaml(n.File)
}

func (n *Network) Load() {
	n.LoadRoute()
	n.LoadOutput()
	n.LoadFindHop()
	n.LoadDNAT()
}

func (n *Network) LoadLink() {
	file := n.Dir("link")
	if err := libol.UnmarshalLoad(&n.Links, file); err != nil {
		libol.Error("Network.LoadLink... %n", err)
	}
}

func (n *Network) LoadRoute() {
	file := n.Dir("route")
	if err := libol.UnmarshalLoad(&n.Routes, file); err != nil {
		libol.Error("Network.LoadRoute... %n", err)
	}
}

func UserShort(value string) string {
	return strings.SplitN(value, "@", 2)[0]
}

func (n *Network) LoadOutput() {
	file := n.Dir("output")
	if err := libol.UnmarshalLoad(&n.Outputs, file); err != nil {
		libol.Error("Network.LoadOutput... %n", err)
	}

	n.LoadLink()
	// Clone link to outputs.
	for _, link := range n.Links {
		link.Correct()
		username := UserShort(link.Username)
		remote, port := SplitSecret(link.Connection)
		value := &Output{
			Protocol: link.Protocol,
			Remote:   remote,
			Secret:   username + ":" + link.Password,
			Crypt:    link.Crypt.Short(),
		}
		fmt.Sscanf(port, "%d", &value.DstPort)
		if _, index := n.FindOutput(value); index == -1 {
			n.Outputs = append(n.Outputs, value)
		}
	}
	n.Links = nil
}

func (n *Network) LoadFindHop() {
	file := n.Dir("findhop")
	if err := libol.UnmarshalLoad(&n.FindHop, file); err != nil {
		libol.Error("Network.LoadFindHop... %n", err)
	}
}

func (n *Network) LoadDNAT() {
	file := n.Dir("dnat")
	if err := libol.UnmarshalLoad(&n.Dnat, file); err != nil {
		libol.Error("Network.LoadDNAT... %n", err)
	}
}

func (n *Network) Save() {
	obj := *n
	obj.Routes = nil
	obj.Outputs = nil
	obj.Dnat = nil
	obj.FindHop = nil // Clear sub dirs.
	if err := libol.MarshalSave(&obj, obj.File, true); err != nil {
		libol.Error("Network.Save %s %s", obj.Name, err)
	}
	n.SaveRoute()
	n.SaveOutput()
	n.SaveFindHop()
	n.SaveDNAT()
}

func (n *Network) SaveRoute() {
	file := n.Dir("route")
	if err := libol.MarshalSave(n.Routes, file, true); err != nil {
		libol.Error("Network.SaveRoute %s %s", n.Name, err)
	}
}

func (n *Network) SaveOutput() {
	file := n.Dir("output")
	if err := libol.MarshalSave(n.Outputs, file, true); err != nil {
		libol.Error("Network.SaveOutput %s %s", n.Name, err)
	}
}

func (n *Network) SaveFindHop() {
	file := n.Dir("findhop")
	if err := libol.MarshalSave(n.FindHop, file, true); err != nil {
		libol.Error("Network.SaveFindHop %s %s", n.Name, err)
	}
}

func (n *Network) SaveDNAT() {
	file := n.Dir("dnat")
	if err := libol.MarshalSave(n.Dnat, file, true); err != nil {
		libol.Error("Network.SaveDNAT %s %s", n.Name, err)
	}
}

func (n *Network) Reload() {
}

func (n *Network) FindRoute(value PrefixRoute) (PrefixRoute, int) {
	for i, obj := range n.Routes {
		if value.Prefix == obj.Prefix {
			return obj, i
		}
	}
	return PrefixRoute{}, -1
}

func (n *Network) ListRoute(call func(value PrefixRoute)) {
	for _, obj := range n.Routes {
		call(obj)
	}
}

func (n *Network) AddRoute(value PrefixRoute) bool {
	_, index := n.FindRoute(value)
	if index == -1 {
		n.Routes = append(n.Routes, value)
	}
	return index == -1
}

func (n *Network) DelRoute(value PrefixRoute) (PrefixRoute, bool) {
	obj, index := n.FindRoute(value)
	if index != -1 {
		n.Routes = append(n.Routes[:index], n.Routes[index+1:]...)
	}
	return obj, index != -1
}

func (n *Network) FindOutput(value *Output) (*Output, int) {
	for i, obj := range n.Outputs {
		if value.Link != "" && value.Link == obj.Link {
			return obj, i
		}
		if value.Link == "" && value.Id() == obj.Id() {
			return obj, i
		}
	}
	return nil, -1
}

func (n *Network) AddOutput(value *Output) bool {
	_, index := n.FindOutput(value)
	if index == -1 {
		n.Outputs = append(n.Outputs, value)
	}
	return index == -1
}

func (n *Network) DelOutput(value *Output) (*Output, bool) {
	obj, index := n.FindOutput(value)
	if index != -1 {
		n.Outputs = append(n.Outputs[:index], n.Outputs[index+1:]...)
	}
	return obj, index != -1
}

func (n *Network) ListOutput(call func(value Output)) {
	for _, obj := range n.Outputs {
		call(*obj)
	}
}

func (n *Network) FindFindHop(value *FindHop) *FindHop {
	if n.FindHop == nil {
		n.FindHop = make(map[string]*FindHop)
	}
	return n.FindHop[value.Name]
}

func (n *Network) AddFindHop(value *FindHop) bool {
	older := n.FindFindHop(value)
	if older == nil {
		n.FindHop[value.Name] = value
		return true
	}
	return false
}

func (n *Network) DelFindHop(value *FindHop) (*FindHop, bool) {
	older := n.FindFindHop(value)
	if older != nil {
		delete(n.FindHop, value.Name)
		return older, true
	}
	return value, false
}

func (n *Network) FindDNAT(value *DNAT) (*DNAT, int) {
	for index, obj := range n.Dnat {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (n *Network) AddDNAT(value *DNAT) bool {
	_, index := n.FindDNAT(value)
	if index == -1 {
		n.Dnat = append(n.Dnat, value)
		return true
	}
	return false
}

func (n *Network) DelDNAT(value *DNAT) (*DNAT, bool) {
	older, index := n.FindDNAT(value)
	if index != -1 {
		n.Dnat = append(n.Dnat[:index], n.Dnat[index+1:]...)
		return older, true
	}
	return older, false
}

func (n *Network) ListDNAT(call func(value DNAT)) {
	for _, obj := range n.Dnat {
		call(*obj)
	}
}
