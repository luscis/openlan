package schema

type VPNClient struct {
	Uptime    int64  `json:"uptime"`
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	Network   string `json:"network"`
	Remote    string `json:"remote"`
	Device    string `json:"device"`
	RxBytes   uint64 `json:"rxBytes"`
	TxBytes   uint64 `json:"txBytes"`
	ErrPkt    uint64 `json:"errors"`
	State     string `json:"state"`
	AliveTime int64  `json:"aliveTime"`
	Address   string `json:"address"`
	System    string `json:"system"`
}

type OpenVPN struct {
	Listen   string   `json:"listen"`
	Protocol string   `json:"protocol,omitempty"`
	Subnet   string   `json:"subnet"`
	Push     []string `json:"push,omitempty"`
}
