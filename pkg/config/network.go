package config

import (
	"github.com/luscis/openlan/pkg/libol"
	"net"
	"path/filepath"
)

type Network struct {
	ConfDir   string        `json:"-"`
	File      string        `json:"file"`
	Alias     string        `json:"-" yaml:"-"`
	Name      string        `json:"name,omitempty" yaml:"name"`
	Provider  string        `json:"provider,omitempty" yaml:"provider"`
	Bridge    *Bridge       `json:"bridge,omitempty" yaml:"bridge,omitempty"`
	Subnet    *IpSubnet     `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	OpenVPN   *OpenVPN      `json:"openvpn,omitempty" yaml:"openvpn,omitempty"`
	Links     []Point       `json:"links,omitempty" yaml:"links,omitempty"`
	Hosts     []HostLease   `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Routes    []PrefixRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
	Acl       string        `json:"acl,omitempty" yaml:"acl,omitempty"`
	Specifies interface{}   `json:"specifies,omitempty" yaml:"specifies,omitempty"`
	Dhcp      string        `json:"dhcp,omitempty" yaml:"dhcp,omitempty"`
	Outputs   []Output      `json:"outputs" yaml:"output,omitempty"`
}

func (n *Network) Correct() {
	if n.Bridge == nil {
		n.Bridge = &Bridge{}
	}
	br := n.Bridge
	br.Network = n.Name
	br.Correct()
	switch n.Provider {
	case "esp":
		spec := n.Specifies
		if obj, ok := spec.(*ESPSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	case "fabric":
		// 28 [udp] - 8 [esp] -
		// 28 [udp] - 8 [vxlan] -
		// 14 [ethernet] - tcp [40] - 1332 [mss] -
		// 42 [padding] ~= variable 30-45
		if br.Mss == 0 {
			br.Mss = 1332
		}
		spec := n.Specifies
		if obj, ok := spec.(*FabricSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	default:
		if n.Subnet == nil {
			n.Subnet = &IpSubnet{}
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
		for i := range n.Routes {
			if n.Routes[i].Metric == 0 {
				n.Routes[i].Metric = 660
			}
			if n.Routes[i].NextHop == "" {
				n.Routes[i].NextHop = ipAddr
			}
			if n.Routes[i].Mode == "" {
				n.Routes[i].Mode = "snat"
			}
		}
		if n.OpenVPN != nil {
			n.OpenVPN.Network = n.Name
			obj := DefaultOpenVPN()
			n.OpenVPN.Correct(obj)
		}
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

func (n *Network) Save() {
	obj := *n
	obj.Routes = nil
	obj.Links = nil
	if err := libol.MarshalSave(&obj, obj.File, true); err != nil {
		libol.Error("Network.Save %s %s", obj.Name, err)
	}
	n.SaveRoute()
	n.SaveLink()
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
