package schema

type Link struct {
	Uptime    int64  `json:"uptime"`
	UUID      string `json:"uuid"`
	Alias     string `json:"alias"`
	Network   string `json:"network"`
	User      string `json:"user"`
	Protocol  string `json:"protocol"`
	Server    string `json:"server"`
	Device    string `json:"device"`
	RxBytes   uint64 `json:"rxBytes"`
	TxBytes   uint64 `json:"txBytes"`
	ErrPkt    uint64 `json:"errors"`
	State     string `json:"state"`
	AliveTime int64  `json:"aliveTime"`
}
