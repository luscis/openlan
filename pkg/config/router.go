package config

type RouterSpecifies struct {
	Mss      int      `json:"tcpMss,omitempty" yaml:"tcpMss,omitempty"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
	Link     string   `json:"link,omitempty" yaml:"link,omitempty"`
	Subnets  []Subnet `json:"subnets,omitempty" yaml:"subnets,omitempty"`
	Loopback string   `json:"loopback,omitempty" yaml:"loopback,omitempty"`
}

func (n *RouterSpecifies) Correct() {
}
