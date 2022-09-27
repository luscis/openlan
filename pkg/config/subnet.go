package config

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

type HostLease struct {
	Network  string `json:"network,omitempty"`
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}
