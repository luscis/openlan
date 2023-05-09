package models

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

type Route struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"nexthop"`
	Metric  int    `json:"metric"`
	Mode    string `json:"mode"`
	Origin  string `json:"origin"`
}

func NewRoute(prefix string, nexthop, mode string) (this *Route) {
	this = &Route{
		Prefix:  prefix,
		NextHop: nexthop,
		Metric:  250,
		Mode:    mode,
	}
	return
}

func (u *Route) String() string {
	return fmt.Sprintf("%s, %s", u.Prefix, u.NextHop)
}

func (u *Route) SetMetric(value int) {
	u.Metric = value
}

func (u *Route) SetOrigin(value string) {
	u.Origin = value
}

type Network struct {
	Name    string   `json:"name"`
	Tenant  string   `json:"tenant,omitempty"`
	IfAddr  string   `json:"ifAddr"`
	IpStart string   `json:"ipStart"`
	IpEnd   string   `json:"ipEnd"`
	Netmask string   `json:"netmask"`
	Routes  []*Route `json:"routes"`
}

func NewNetwork(name string, ifAddr string) (this *Network) {
	address := ifAddr
	netmask := "255.255.255.255"
	s := strings.SplitN(ifAddr, "/", 2)
	if len(s) == 2 {
		address = s[0]
		_, n, err := net.ParseCIDR(ifAddr)
		if err == nil {
			netmask = net.IP(n.Mask).String()
		} else {
			netmask = s[1]
		}
	}
	this = &Network{
		Name:    name,
		IfAddr:  address,
		Netmask: netmask,
	}
	return
}

func (u *Network) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s, %s",
		u.Name, u.IfAddr, u.IpStart, u.IpEnd, u.Netmask, u.Routes)
}

func (u *Network) ParseIP(s string) {
}

func NetworkEqual(o *Network, n *Network) bool {
	if o == n {
		return true
	} else if o == nil || n == nil {
		return false
	} else if o.IfAddr != n.IfAddr || o.Netmask != n.Netmask {
		return false
	} else {
		ors := make([]string, 0, 32)
		nrs := make([]string, 0, 32)
		for _, rt := range o.Routes {
			ors = append(ors, rt.String())
		}
		for _, rt := range n.Routes {
			nrs = append(nrs, rt.String())
		}
		if len(ors) != len(nrs) {
			return false
		}
		sort.Strings(ors)
		sort.Strings(nrs)
		for i := range ors {
			if ors[i] != nrs[i] {
				return false
			}
		}
		return true
	}
}
