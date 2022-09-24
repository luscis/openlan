package config

type IpSubnet struct {
	Network string `json:"network,omitempty" yaml:"network,omitempty"`
	Start   string `json:"start,omitempty" yaml:"start,omitempty"`
	End     string `json:"end,omitempty" yaml:"end,omitempty"`
	Netmask string `json:"netmask,omitempty" yaml:"netmask,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Weight  int    `json:"weight"`
}

type PrefixRoute struct {
	File      string      `json:"-" yaml:"-"`
	Network   string      `json:"network,omitempty" yaml:"network,omitempty"`
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop"`
	MultiPath []MultiPath `json:"multipath,omitempty"`
	Metric    int         `json:"metric"`
	Mode      string      `json:"forward,omitempty" yaml:"forward,omitempty"` // route or snat
}

type HostLease struct {
	Network  string `json:"network,omitempty" yaml:"network,omitempty"`
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}
