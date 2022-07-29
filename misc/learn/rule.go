package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	//"net"
)

func main() {
	rules, err := netlink.RuleList(netlink.FAMILY_V4)
	if err != nil {
		panic(err)
	}
	for _, ru := range rules {
		fmt.Println(ru)
	}
	ru := netlink.NewRule()
	//src := &net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.CIDRMask(0, 32)}
	ru.Table = 100
	ru.Priority = 16383
	if err := netlink.RuleAdd(ru); err != nil {
		fmt.Printf("%s %s\n", ru, err)
	}
}
