package schema

import "fmt"

type Lease struct {
	Address string `json:"address"`
	Alias   string `json:"alias"`
	Client  string `json:"client"`
	Type    string `json:"type"`
	Network string `json:"network"`
}

type PrefixRoute struct {
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop,omitempty"`
	FindHop   string      `json:"findhop,omitempty"`
	Metric    int         `json:"metric"`
	Link      string      `json:"link,omitempty"`
	Table     int         `json:"table,omitempty"`
	Source    string      `json:"source,omitempty"`
	Protocol  string      `json:"protocol,omitempty"`
	MultiPath []MultiPath `json:"multipath,omitempty"`
}

type KernelRoute struct {
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop,omitempty"`
	Metric    int         `json:"metric"`
	Link      string      `json:"link,omitempty"`
	Table     int         `json:"table,omitempty"`
	Source    string      `json:"source,omitempty"`
	Protocol  string      `json:"protocol,omitempty"`
	Multipath []MultiPath `json:"multipath,omitempty"`
}

type RedirectRoute struct {
	Source  string `json:"source,omitempty"`
	Table   int    `json:"table,omitempty"`
	NextHop string `json:"nexthop,omitempty"`
}

func (r RedirectRoute) Rule() string {
	return fmt.Sprintf("source:%s lookup:%d", r.Source, r.Table)
}

func (r RedirectRoute) Route() string {
	return fmt.Sprintf(" table:%d nexthop:%s", r.Table, r.NextHop)
}

func (k KernelRoute) ID() string {
	return fmt.Sprintf("%d-%s-%s", k.Table, k.Protocol, k.Prefix)
}

type KernelNeighbor struct {
	Link    string `json:"link,omitempty"`
	Address string `json:"address,omitempty"`
	HwAddr  string `json:"hwaddr,omitempty"`
	State   string `json:"state,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Link    string `json:"link"`
	Weight  int    `json:"weight"`
}

type Subnet struct {
	IfAddr  string `json:"address,omitempty"`
	IpStart string `json:"startAt,omitempty"`
	IpEnd   string `json:"endAt,omitempty"`
	Netmask string `json:"netmask"`
}

type Network struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type FindHop struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	Check     string `json:"check"`
	NextHop   string `json:"nexthop"`
	Available string `json:"available"`
}

type SNAT struct {
	Scope string `json:"scope"`
}

type DNAT struct {
	Protocol string `json:"protocol"`
	Dest     string `json:"destination,omitempty"`
	Dport    int    `json:"dport"`
	ToDest   string `json:"todestination"`
	ToDport  int    `json:"todport"`
}

type RouterTunnel struct {
	Protocol string `json:"protocol"`
	Remote   string `json:"remote"`
	Address  string `json:"address"`
}

type RouterPrivate struct {
	Subnet string `json:"subnet"`
}

type IPAddress struct {
	Address string `json:"address"`
}

type RouterInterface struct {
	Device  string `json:"device"`
	VLAN    int    `json:"vlan"`
	Address string `json:"address"`
}

type KernelUsage struct {
	CPUUsage  int    `json:"cpuUsage"`
	MemUsed   uint64 `json:"memUsed"`
	MemTotal  uint64 `json:"memTotal"`
	DiskUsed  uint64 `json:"diskUsed"`
	DiskTotal uint64 `json:"diskTotal"`
}

type KernelConntrack struct {
	Total int `json:"total"`
	TCP   int `json:"tcp"`
	UDP   int `json:"udp"`
	ICMP  int `json:"icmp"`
}
