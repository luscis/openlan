package config

type Dhcp struct {
	Name   string        `json:"name,omitempty"`
	Bridge *Bridge       `json:"bridge,omitempty"`
	Subnet *Subnet       `json:"subnet,omitempty"`
	Hosts  []HostLease   `json:"hosts,omitempty"`
	Routes []PrefixRoute `json:"routes,omitempty"`
}
