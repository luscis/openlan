package config

type L2TP struct {
	Address string   `json:"address"`
	Subnet  *Subnet  `json:"subnet,omitempty"`
	Options []string `json:"Options,omitempty"`
	IpSec   string   `json:"ipsec,omitempty"`
}
