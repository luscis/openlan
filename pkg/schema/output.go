package schema

type Output struct {
	Network   string `json:"network"`
	Protocol  string `json:"protocol"`
	Remote    string `json:"remote"`
	DstPort   int    `json:"dstPort"`
	Segment   int    `json:"segment"`
	Secret    string `json:"secret"`
	Crypt     string `json:"crypt"`
	Device    string `json:"device"`
	RxBytes   uint64 `json:"rxBytes"`
	TxBytes   uint64 `json:"txBytes"`
	ErrPkt    uint64 `json:"errors"`
	AliveTime int64  `json:"aliveTime"`
}
