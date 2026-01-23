package schema

type Access struct {
	Uptime    int64  `json:"uptime"`
	UUID      string `json:"uuid"`
	Network   string `json:"network,omitempty"`
	User      string `json:"user"`
	Alias     string `json:"alias,omitempty"`
	Protocol  string `json:"protocol"`
	Remote    string `json:"remote"`
	Device    string `json:"device"`
	RxBytes   uint64 `json:"rxBytes"`
	TxBytes   uint64 `json:"txBytes"`
	ErrPkt    uint64 `json:"errors,omitempty"`
	State     string `json:"state"`
	AliveTime int64  `json:"aliveTime"`
	System    string `json:"system,omitempty"`
	Address   string `json:"address,omitempty"`
	Fallback  string `json:"fallback,omitempty"`
}
