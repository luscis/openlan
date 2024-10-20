package config

import (
	"net"
	"path/filepath"

	"github.com/luscis/openlan/pkg/libol"
)

type Network struct {
	ConfDir   string              `json:"-"`
	File      string              `json:"-"`
	Alias     string              `json:"-"`
	Name      string              `json:"name"`
	Provider  string              `json:"provider,omitempty"`
	Bridge    *Bridge             `json:"bridge,omitempty"`
	Subnet    *Subnet             `json:"subnet,omitempty"`
	OpenVPN   *OpenVPN            `json:"openvpn,omitempty"`
	Links     []Point             `json:"links,omitempty"`
	Hosts     []HostLease         `json:"hosts,omitempty"`
	Routes    []PrefixRoute       `json:"routes,omitempty"`
	Acl       string              `json:"acl,omitempty"`
	Specifies interface{}         `json:"specifies,omitempty"`
	Dhcp      string              `json:"dhcp,omitempty"`
	Outputs   []*Output           `json:"outputs,omitempty"`
	ZTrust    string              `json:"ztrust,omitempty"`
	Qos       string              `json:"qos,omitempty"`
	Snat      string              `json:"snat,omitempty"`
	Namespace string              `json:"namespace,omitempty"`
	FindHop   map[string]*FindHop `json:"findhop,omitempty"`
}

func (n *Network) NewSpecifies() interface{} {
	switch n.Provider {
	case "ipsec":
		n.Specifies = &IPSecSpecifies{}
	case "router":
		n.Specifies = &RouterSpecifies{}
	default:
		n.Specifies = nil
	}
	return n.Specifies
}

func (n *Network) Correct(sw *Switch) {
	ipAddr := ""
	ipMask := ""
	if n.Bridge == nil {
		n.Bridge = &Bridge{}
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
	default:
		br := n.Bridge
		br.Network = n.Name
		br.Correct()
		if _i, _n, err := net.ParseCIDR(br.Address); err == nil {
			ipAddr = _i.String()
			ipMask = net.IP(_n.Mask).String()
		}
	}
	if n.Subnet == nil {
		n.Subnet = &Subnet{}
	}
	if n.Subnet.Netmask == "" {
		n.Subnet.Netmask = ipMask
	}
	CorrectRoutes(n.Routes, ipAddr)
	if n.OpenVPN != nil {
		n.OpenVPN.Network = n.Name
		obj := DefaultOpenVPN()
		n.OpenVPN.Merge(obj)
		n.OpenVPN.Correct(sw)
	}

	for key, value := range n.FindHop {
		value.Correct()
		n.FindHop[key] = value
	}
}

func (n *Network) Dir(elem ...string) string {
	args := append([]string{n.ConfDir}, elem...)
	return filepath.Join(args...)
}

func (n *Network) LoadLink() {
	file := n.Dir("link", n.Name+".json")
	if err := libol.UnmarshalLoad(&n.Links, file); err != nil {
		libol.Error("Network.LoadLink... %n", err)
	}
}

func (n *Network) LoadRoute() {
	file := n.Dir("route", n.Name+".json")
	if err := libol.UnmarshalLoad(&n.Routes, file); err != nil {
		libol.Error("Network.LoadRoute... %n", err)
	}
}

func (n *Network) LoadOutput() {
	file := n.Dir("output", n.Name+".json")
	if err := libol.UnmarshalLoad(&n.Outputs, file); err != nil {
		libol.Error("Network.LoadOutput... %n", err)
	}
}

func (n *Network) LoadFindHop() {
	file := n.Dir("findhop", n.Name+".json")
	if err := libol.UnmarshalLoad(&n.FindHop, file); err != nil {
		libol.Error("Network.LoadFindHop... %n", err)
	}
}

func (n *Network) Save() {
	obj := *n
	obj.Routes = nil
	obj.Links = nil
	obj.Outputs = nil
	if err := libol.MarshalSave(&obj, obj.File, true); err != nil {
		libol.Error("Network.Save %s %s", obj.Name, err)
	}
	n.SaveRoute()
	n.SaveLink()
	n.SaveOutput()
	n.SaveFindHop()
}

func (n *Network) SaveRoute() {
	file := n.Dir("route", n.Name+".json")
	if n.Routes == nil {
		return
	}
	if err := libol.MarshalSave(n.Routes, file, true); err != nil {
		libol.Error("Network.SaveRoute %s %s", n.Name, err)
	}
}

func (n *Network) SaveLink() {
	file := n.Dir("link", n.Name+".json")
	if n.Links == nil {
		return
	}
	if err := libol.MarshalSave(n.Links, file, true); err != nil {
		libol.Error("Network.SaveLink %s %s", n.Name, err)
	}
}

func (n *Network) SaveOutput() {
	file := n.Dir("output", n.Name+".json")
	if n.Outputs == nil {
		return
	}
	if err := libol.MarshalSave(n.Outputs, file, true); err != nil {
		libol.Error("Network.SaveOutput %s %s", n.Name, err)
	}
}

func (n *Network) SaveFindHop() {
	file := n.Dir("findhop", n.Name+".json")
	if n.FindHop == nil {
		return
	}
	if err := libol.MarshalSave(n.FindHop, file, true); err != nil {
		libol.Error("Network.SaveFindHop %s %s", n.Name, err)
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
