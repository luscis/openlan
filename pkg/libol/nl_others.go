//go:build !linux
// +build !linux

package libol

import "net"

func GetLocalByGw(addr string) (net.IP, error) {
	return nil, NewErr("GetLocalByGw notSupport")
}

func ListRoutes() ([]Prefix, error) {
	return nil, NewErr("ListRoutes notSupport")
}

func ListNeighbrs() ([]Neighbor, error) {
	return nil, NewErr("ListNeighbors notSupport")
}

func ListConnStats() ConnStats {
	return ConnStats{}
}

func ListPhyLinks() []Device {
	return nil
}
