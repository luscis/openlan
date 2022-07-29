package config

type Output struct {
	Vlan      int    `json:"vlan"`
	Interface string `json:"interface"` // format, like: gre:<addr>, vxlan:<addr>:<vni>
	Link      string `json:"link"`      // link name
}
