package config

type VxLANSpecifies struct {
	Name   string `json:"name"`
	Vni    uint32 `json:"vni"`
	Fabric string `json:"fabric"`
}

func (c *VxLANSpecifies) Correct() {
}
