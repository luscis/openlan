package config

type Dhcp struct {
	Name      string        `json:"name,omitempty" yaml:"name,omitempty"`
	Interface string        `json:"interface,omitempty" yaml:"interface,omitempty"`
	Subnet    *Subnet       `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Hosts     []HostLease   `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Routes    []PrefixRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
}
