package schema

import "fmt"

type Device struct {
	Network string `json:"network,omitempty"`
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
	Mac     string `json:"mac,omitempty"`
	Mtu     int    `json:"mtu,omitempty"`
	Send    uint64 `json:"send,omitempty"`
	Recv    uint64 `json:"recv,omitempty"`
	Drop    uint64 `json:"drop,omitempty"`
	RxSpeed uint64 `json:"rxSpeed,omitempty"`
	TxSpeed uint64 `json:"txSpeed,omitempty"`
}

func (c Device) ID() string {
	return fmt.Sprintf("%s:%s", c.Network, c.Name)
}

type HwMacInfo struct {
	Uptime  int64  `json:"uptime"`
	Address string `json:"address"`
	Device  string `json:"device"`
}
