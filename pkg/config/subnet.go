package config

import (
	"fmt"
	"strings"
)

type Subnet struct {
	Network string `json:"network,omitempty"`
	Start   string `json:"startAt,omitempty"`
	End     string `json:"endAt,omitempty"`
	Netmask string `json:"netmask,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Weight  int    `json:"weight"`
}

type PrefixRoute struct {
	File      string      `json:"-"`
	Network   string      `json:"network,omitempty"`
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop"`
	MultiPath []MultiPath `json:"multipath,omitempty"`
	Metric    int         `json:"metric"`
	Mode      string      `json:"forward,omitempty"` // route or snat
}

func (r *PrefixRoute) String() string {
	elems := []string{}
	if len(r.Prefix) > 0 {
		elems = append(elems, fmt.Sprintf("Prefix: %s", r.Prefix))
	}
	if len(r.NextHop) > 0 {
		elems = append(elems, fmt.Sprintf("Nexthop: %s", r.NextHop))
	}
	if len(r.Mode) > 0 {
		elems = append(elems, fmt.Sprintf("Forward: %s", r.Mode))
	}
	if r.Metric > 0 {
		elems = append(elems, fmt.Sprintf("Metric: %d", r.Metric))
	}
	return fmt.Sprintf("{%s}", strings.Join(elems, " "))
}

type HostLease struct {
	Network  string `json:"network,omitempty"`
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}
