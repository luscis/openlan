package config

import (
	"net"
	"path/filepath"

	"github.com/luscis/openlan/pkg/libol"
)

type Network struct {
	ConfDir   string               `json:"-"`
	File      string               `json:"file"`
	Alias     string               `json:"-"`
	Name      string               `json:"name"`
	Provider  string               `json:"provider,omitempty"`
	Bridge    *Bridge              `json:"bridge,omitempty"`
	Subnet    *Subnet              `json:"subnet,omitempty"`
	OpenVPN   *OpenVPN             `json:"openvpn,omitempty"`
	Links     []Point              `json:"links,omitempty"`
	Hosts     []HostLease          `json:"hosts,omitempty"`
	Routes    []PrefixRoute        `json:"routes,omitempty"`
	Acl       string               `json:"acl,omitempty"`
	Specifies interface{}          `json:"specifies,omitempty"`
	Dhcp      string               `json:"dhcp,omitempty"`
	Outputs   []Output             `json:"outputs,omitempty"`
	ZTrust    string               `json:"ztrust,omitempty"`
	Qos       string               `json:"qos,omitempty"`
	Namespace string               `json:"namespace,omitempty"`
	NextGroup map[string]NextGroup `json:"nextgroup,omitempty"`
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
	if n.Bridge == nil {
		n.Bridge = &Bridge{}
	}
	br := n.Bridge
	br.Network = n.Name
	br.Correct()
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
	}
	if n.Subnet == nil {
		n.Subnet = &Subnet{}
	}
	ipAddr := ""
	ipMask := ""
	if _i, _n, err := net.ParseCIDR(br.Address); err == nil {
		ipAddr = _i.String()
		ipMask = net.IP(_n.Mask).String()
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

	for key, value := range n.NextGroup {
		value.Correct()
		n.NextGroup[key] = value
	}
}

func (n *Network) Dir(elem ...string) string {
	args := append([]string{n.ConfDir}, elem...)
	return filepath.Join(args...)
}

func (n *Network) LoadLink() {
	file := n.Dir("link", n.Name+".json")
	if err := libol.FileExist(file); err == nil {
		if err := libol.UnmarshalLoad(&n.Links, file); err != nil {
			libol.Error("Network.LoadLink... %n", err)
		}
	}
}

func (n *Network) LoadRoute() {
	file := n.Dir("route", n.Name+".json")
	if err := libol.FileExist(file); err == nil {
		if err := libol.UnmarshalLoad(&n.Routes, file); err != nil {
			libol.Error("Network.LoadRoute... %n", err)
		}
	}
}

func (n *Network) LoadOutput() {
	file := n.Dir("output", n.Name+".json")
	if err := libol.FileExist(file); err == nil {
		if err := libol.UnmarshalLoad(&n.Outputs, file); err != nil {
			libol.Error("Network.LoadOutput... %n", err)
		}
	}
}

func (n *Network) LoadNextGroup() {
	file := n.Dir("nextgroup", n.Name+".json")
	if err := libol.FileExist(file); err == nil {
		if err := libol.UnmarshalLoad(&n.NextGroup, file); err != nil {
			libol.Error("Network.LoadNextGroup... %n", err)
		}
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
	n.SaveNextGroup()
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

func (n *Network) SaveNextGroup() {
	file := n.Dir("nextgroup", n.Name+".json")
	if n.NextGroup == nil {
		return
	}
	if err := libol.MarshalSave(n.NextGroup, file, true); err != nil {
		libol.Error("Network.SaveNextGroup %s %s", n.Name, err)
	}
}

func (n *Network) Reload() {
}
