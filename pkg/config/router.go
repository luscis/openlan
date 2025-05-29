package config

type RouterSpecifies struct {
	Mss       int      `json:"tcpMss,omitempty" yaml:"tcpMss,omitempty"`
	Name      string   `json:"-" yaml:"-"`
	Link      string   `json:"link,omitempty" yaml:"link,omitempty"`
	Subnets   []Subnet `json:"subnets,omitempty" yaml:"subnets,omitempty"`
	Loopback  string   `json:"loopback,omitempty" yaml:"loopback,omitempty"`
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
}

func (n *RouterSpecifies) Correct() {
}
