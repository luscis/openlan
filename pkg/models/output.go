package models

import (
	"fmt"
	"os"
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
	Crypt     string
	RxBytes   uint64
	TxBytes   uint64
	ErrPkt    uint64
	NewTime   int64
	Fallback  string
	StatsFile string
	PidFile   string
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
		if !o.hasLiveProcess() {
			o.uptime = 0
			return "down"
		}
		sts := &schema.Access{}
		_ = libol.UnmarshalLoad(sts, o.StatsFile)
		o.uptime = sts.AliveTime
		o.Device = sts.Device
		return sts.State
	}
	return ""
}

func (o *Output) hasLiveProcess() bool {
	if o.PidFile == "" {
		return true
	}
	data, err := os.ReadFile(o.PidFile)
	if err != nil {
		return false
	}
	pid := 0
	_, _ = fmt.Sscanf(string(data), "%d", &pid)
	return pid > 0 && libol.HasProcess(pid)
}
