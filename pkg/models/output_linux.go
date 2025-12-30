package models

import "github.com/luscis/openlan/pkg/network"

func (l *Output) Update() {
	sts := network.GetStats(l.Device)
	l.RxBytes = sts.Recv
	l.TxBytes = sts.Send
}
