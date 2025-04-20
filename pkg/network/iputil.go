package network

import "net"

func LookupIP(name string) string {
	if addr, _ := net.LookupIP(name); len(addr) > 0 {
		return addr[0].String()
	}
	return ""
}
