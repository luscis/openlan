package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os"
)

func main() {
	dest_str := os.Getenv("DEST")
	dest := net.ParseIP(dest_str)
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		panic(err)
	}
	var hit *net.IPNet
	for _, rte := range routes {
		fmt.Println(rte)
		if rte.Dst != nil && !rte.Dst.Contains(dest) {
			continue
		}
		if hit != nil {
			rts, _ := rte.Dst.Mask.Size()
			ths, _ := hit.Mask.Size()
			if rts < ths {
				continue
			}
		}
		hit = rte.Dst
		ifIndex := rte.LinkIndex
		gateway := rte.Gw
		if gateway == nil {
			gateway = rte.Src
		}
		fmt.Println("gw", rte.Gw)
		link, _ := netlink.LinkByIndex(ifIndex)
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			if addr.Contains(gateway) {
				fmt.Println("hit ", addr.IP)
			}
		}
	}
}
