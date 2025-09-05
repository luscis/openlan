//go:build !linux
// +build !linux

package libol

import "net"

func GetLocalByGw(addr string) (net.IP, error) {
	return nil, NewErr("GetLocalByGw notSupport")
}

func ListRoutes() ([]Prefix, error) {
	return nil, NewErr("ListRoute notSupport")
}
