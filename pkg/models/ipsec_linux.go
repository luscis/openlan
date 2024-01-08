package models

import (
	"github.com/luscis/openlan/pkg/libol"
	nl "github.com/vishvananda/netlink"
	"time"
)

func (l *EspState) Update() {
	used := int64(0)
	if xss, err := nl.XfrmStateGet(l.In); xss != nil {
		l.TxBytes = xss.Statistics.Bytes
		l.TxPackages = xss.Statistics.Packets
		used = int64(xss.Statistics.UseTime)
	} else {
		libol.Debug("EspState.Update %s", err)
	}

	if xss, err := nl.XfrmStateGet(l.Out); xss != nil {
		l.RxBytes = xss.Statistics.Bytes
		l.RxPackages = xss.Statistics.Packets
	} else {
		libol.Debug("EspState.Update %s", err)
	}

	if used > 0 {
		l.AliveTime = time.Now().Unix() - used
	}
}
