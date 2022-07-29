package ss

import "time"

var config struct {
	Verbose    bool
	UDPTimeout time.Duration
	TCPCork    bool
}
