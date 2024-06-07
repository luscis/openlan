package models

import "time"

type Output struct {
	Network  string
	Protocol string
	Remote   string
	Segment  int
	Device   string
	Secret   string
	RxBytes  uint64
	TxBytes  uint64
	ErrPkt   uint64
	NewTime  int64
}

func (o *Output) UpTime() int64 {
	return time.Now().Unix() - o.NewTime
}
