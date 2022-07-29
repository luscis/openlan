package config

type Dhcp struct {
	Name   string        `json:"name,omitempty" yaml:"name"`
	Bridge *Bridge       `json:"bridge,omitempty" yaml:"bridge,omitempty"`
	Subnet *IpSubnet     `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Hosts  []HostLease   `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Routes []PrefixRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
}
