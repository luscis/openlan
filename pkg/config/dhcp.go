package config

type Dhcp struct {
	Name      string        `json:"name,omitempty"`
	Interface string        `json:"interface,omitempty"`
	Subnet    *Subnet       `json:"subnet,omitempty"`
	Hosts     []HostLease   `json:"hosts,omitempty"`
	Routes    []PrefixRoute `json:"routes,omitempty"`
}
