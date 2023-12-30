package cswitch

import (
	"fmt"
	"net"
	"testing"
)

func TestDNS_lookup(t *testing.T) {
	addr, err := net.LookupHost("nj.openlan.net")
	fmt.Println(addr, err)
	addr, err = net.LookupHost("114.221.197.118")
	fmt.Println(addr, err)
}
