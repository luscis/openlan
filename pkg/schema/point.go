package schema

type Point struct {
	Uptime    int64   `json:"uptime"`
	UUID      string  `json:"uuid"`
	Network   string  `json:"network"`
	User      string  `json:"user"`
	Alias     string  `json:"alias"`
	Protocol  string  `json:"protocol"`
	Remote    string  `json:"remote"`
	Switch    string  `json:"switch,omitempty"`
	Device    string  `json:"device"`
	RxBytes   int64   `json:"rxBytes"`
	TxBytes   int64   `json:"txBytes"`
	ErrPkt    int64   `json:"errors"`
	State     string  `json:"state"`
	AliveTime int64   `json:"aliveTime"`
	System    string  `json:"system"`
	Address   Network `json:"address"`
}
