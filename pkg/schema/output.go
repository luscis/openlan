package schema

type Output struct {
	Network   string `json:"network"`
	Protocol  string `json:"protocol"`
	Remote    string `json:"remote"`
	DstPort   int    `json:"dstPort,omitempty"`
	Segment   int    `json:"segment,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Crypt     string `json:"crypt,omitempty"`
	Device    string `json:"device"`
	RxBytes   uint64 `json:"rxBytes,omitempty"`
	TxBytes   uint64 `json:"txBytes,omitempty"`
	ErrPkt    uint64 `json:"errors,omitempty"`
	AliveTime int64  `json:"aliveTime"`
	Fallback  string `json:"fallback,omitempty"`
}
