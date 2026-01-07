package models

import (
	"time"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type Output struct {
	Network   string
	Protocol  string
	Remote    string
	Segment   int
	Device    string
	Secret    string
	RxBytes   uint64
	TxBytes   uint64
	ErrPkt    uint64
	NewTime   int64
	Fallback  string
	StatsFile string
	uptime    int64
}

func (o *Output) UpTime() int64 {
	if o.uptime > 0 {
		return o.uptime
	}
	return time.Now().Unix() - o.NewTime
}

func (o *Output) GetState() string {
	if o.StatsFile != "" {
		sts := &schema.Access{}
		_ = libol.UnmarshalLoad(sts, o.StatsFile)
		o.uptime = sts.AliveTime
		return sts.State
	}
	return ""
}
