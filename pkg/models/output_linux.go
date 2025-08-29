package models

import (
	nl "github.com/vishvananda/netlink"
)

func (l *Output) Update() {
	if link, err := nl.LinkByName(l.Device); err == nil {
		sts := link.Attrs().Statistics
		l.RxBytes = sts.RxBytes
		l.TxBytes = sts.TxBytes
	}
}
