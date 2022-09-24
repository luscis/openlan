package config

type VxLANSpecifies struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Vni    uint32 `json:"vni"`
	Fabric string `json:"fabric"`
}

func (c *VxLANSpecifies) Correct() {
}
