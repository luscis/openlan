package schema

type VPNClient struct {
	Uptime    int64  `json:"uptime"`
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	Network   string `json:"network"`
	User      string `json:"user"`
	Remote    string `json:"remote"`
	Device    string `json:"device"`
	RxBytes   int64  `json:"rxBytes"`
	TxBytes   int64  `json:"txBytes"`
	ErrPkt    string `json:"errors"`
	State     string `json:"state"`
	AliveTime int64  `json:"aliveTime"`
	Address   string `json:"address"`
}
