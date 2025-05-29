package config

type VxLANSpecifies struct {
	Name   string `json:"-" yaml:"-"`
	Vni    uint32 `json:"vni" yaml:"vni"`
	Fabric string `json:"fabric" yaml:"fabric"`
}

func (c *VxLANSpecifies) Correct() {
}
