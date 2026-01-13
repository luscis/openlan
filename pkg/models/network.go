package models

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

type Route struct {
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop"`
	Metric    int         `json:"metric"`
	Origin    string      `json:"origin"`
	MultiPath []MultiPath `json:"multipath,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Weight  int    `json:"weight"`
}

func NewRoute(prefix, nexthop string) (this *Route) {
	this = &Route{
		Prefix:  prefix,
		NextHop: nexthop,
		Metric:  250,
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
	Name    string      `json:"name"`
	Tenant  string      `json:"tenant,omitempty"`
	Gateway string      `json:"gateway,omitempty"`
	Address string      `json:"address,omitempty"`
	IpStart string      `json:"startAt,omitempty"`
	IpEnd   string      `json:"endAt,omitempty"`
	Netmask string      `json:"netmask,omitempty"`
	Routes  []*Route    `json:"routes,omitempty"`
	Config  interface{} `json:"config,omitempty"`
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
		Address: address,
		Netmask: netmask,
	}
	return
}

func (n *Network) String() string {
	return fmt.Sprintf("name:%s gateway:%s address:%s netmask:%s routes:%v", n.Name, n.Gateway, n.Address, n.Netmask, n.Routes)
}

func (u *Network) ParseIP(s string) {
}

func NetworkEqual(o *Network, n *Network) bool {
	if o == n {
		return true
	} else if o == nil || n == nil {
		return false
	} else if o.Address != n.Address || o.Netmask != n.Netmask {
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
