// +build !linux

package libol

import "net"

func GetLocalByGw(addr string) (net.IP, error) {
	return nil, NewErr("GetLocalByGw notSupport")
}
