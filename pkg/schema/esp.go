package schema

import "net"

type Esp struct {
	Name    string      `json:"name"`
	Address string      `json:"address"`
	Members []EspMember `json:"members,omitempty"`
}

type EspState struct {
	Name       string `json:"name"`
	AliveTime  int64  `json:"alive"`
	Spi        int    `json:"spi"`
	Local      net.IP `json:"source"`
	Mode       uint8  `json:"mode"`
	Proto      uint8  `json:"proto"`
	Remote     net.IP `json:"destination"`
	Auth       string `json:"auth"`
	Crypt      string `json:"crypt"`
	Encap      string `json:"encap" `
	RemotePort int    `json:"remotePort"`
	TxBytes    int64  `json:"txBytes"`
	TxPackages int64  `json:"txPackages"`
	RxBytes    int64  `json:"rxBytes"`
	RxPackages int64  `json:"rxPackages"`
}

type EspPolicy struct {
	Name   string `json:"name"`
	Spi    int    `json:"spi"`
	Local  net.IP `json:"local"`
	Remote net.IP `json:"remote"`
	Source string `json:"source"`
	Dest   string `json:"destination"`
}

type EspMember struct {
	Name   string      `json:"name"`
	Spi    uint32      `json:"spi"`
	Peer   string      `json:"peer"`
	State  EspState    `json:"state"`
	Policy []EspPolicy `json:"policy"`
}
