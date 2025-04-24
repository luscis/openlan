package config

type VxLANSpecifies struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Vni    uint32 `json:"vni" yaml:"vni"`
	Fabric string `json:"fabric" yaml:"fabric"`
}

func (c *VxLANSpecifies) Correct() {
}
