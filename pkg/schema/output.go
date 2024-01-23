package schema

type Output struct {
	Network   string `json:"network"`
	Protocol  string `json:"protocol"`
	Remote    string `json:"Remote"`
	Segment   int    `json:"segment"`
	Device    string `json:"device"`
	RxBytes   uint64 `json:"rxBytes"`
	TxBytes   uint64 `json:"txBytes"`
	ErrPkt    uint64 `json:"errors"`
	AliveTime int64  `json:"aliveTime"`
}
