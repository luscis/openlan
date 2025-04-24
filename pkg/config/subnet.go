package config

import (
	"fmt"
	"strings"
)

type Subnet struct {
	Network string `json:"network,omitempty" yaml:"network,omitempty"`
	Start   string `json:"startAt,omitempty" yaml:"startAt,omitempty"`
	End     string `json:"endAt,omitempty" yaml:"endAt,omitempty"`
	Netmask string `json:"netmask,omitempty" yaml:"netmask,omitempty"`
	CIDR    string `json:"cidr,omitempty" yaml:"cidr,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop" yaml:"nexthop"`
	Weight  int    `json:"weight" yaml:"weight"`
}

func (mp *MultiPath) CompareEqual(b MultiPath) bool {
	return mp.NextHop == b.NextHop
}

type PrefixRoute struct {
	File      string      `json:"-" yaml:"-"`
	Network   string      `json:"network,omitempty" yaml:"network,omitempty"`
	Prefix    string      `json:"prefix" yaml:"prefix"`
	NextHop   string      `json:"nexthop" yaml:"nexthop"`
	MultiPath []MultiPath `json:"multipath,omitempty" yaml:"multipath,omitempty"`
	Metric    int         `json:"metric" yaml:"metric"`
	FindHop   string      `json:"findhop,omitempty" yaml:"findhop,omitempty"`
}

func (r *PrefixRoute) String() string {
	elems := []string{}
	if len(r.Prefix) > 0 {
		elems = append(elems, fmt.Sprintf("Prefix: %s", r.Prefix))
	}
	if len(r.NextHop) > 0 {
		elems = append(elems, fmt.Sprintf("Nexthop: %s", r.NextHop))
	}
	if len(r.FindHop) > 0 {
		elems = append(elems, fmt.Sprintf("Findhop: %s", r.FindHop))
	}
	if r.Metric > 0 {
		elems = append(elems, fmt.Sprintf("Metric: %d", r.Metric))
	}
	return fmt.Sprintf("{%s}", strings.Join(elems, " "))
}

func (r *PrefixRoute) CorrectRoute(nexthop string) {
	if r.Metric == 0 {
		r.Metric = 660
	}
	if r.NextHop == "" {
		r.NextHop = nexthop
	}
}

func CorrectRoutes(routes []PrefixRoute, nexthop string) {
	for i := range routes {
		routes[i].CorrectRoute(nexthop)
	}
}

type HostLease struct {
	Network  string `json:"network,omitempty" yaml:"network,omitempty"`
	Hostname string `json:"hostname" yaml:"hostname"`
	Address  string `json:"address" yaml:"address"`
}
