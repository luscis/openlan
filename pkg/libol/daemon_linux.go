package libol

import "github.com/coreos/go-systemd/v22/daemon"

func PreNotify() {
}

func SdNotify() {
	go daemon.SdNotify(false, daemon.SdNotifyReady)
}
