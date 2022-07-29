package config

type IpSubnet struct {
	Network string `json:"network,omitempty"`
	Start   string `json:"start,omitempty"`
	End     string `json:"end,omitempty"`
	Netmask string `json:"netmask,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Weight  int    `json:"weight"`
}

type PrefixRoute struct {
	File      string      `json:"file,omitempty"`
	Network   string      `json:"network,omitempty"`
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop"`
	MultiPath []MultiPath `json:"multipath,omitempty"`
	Metric    int         `json:"metric"`
	Mode      string      `json:"mode" yaml:"forwardMode"` // route or snat
}

type HostLease struct {
	Network  string `json:"network"`
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}
