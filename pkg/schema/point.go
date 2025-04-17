package schema

type Point struct {
	Uptime    int64             `json:"uptime"`
	UUID      string            `json:"uuid"`
	Network   string            `json:"network"`
	User      string            `json:"user"`
	Alias     string            `json:"alias"`
	Protocol  string            `json:"protocol"`
	Remote    string            `json:"remote"`
	Device    string            `json:"device"`
	RxBytes   uint64            `json:"rxbytes"`
	TxBytes   uint64            `json:"txbytes"`
	ErrPkt    uint64            `json:"errors"`
	State     string            `json:"state"`
	AliveTime int64             `json:"alivetime"`
	System    string            `json:"system"`
	Address   string            `json:"address"`
	Names     map[string]string `json:"names"`
}
