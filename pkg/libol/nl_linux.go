package libol

import (
	"github.com/vishvananda/netlink"
	"net"
)

func GetLocalByGw(addr string) (net.IP, error) {
	local := net.IP{}
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	dest := net.ParseIP(addr)
	if dest == nil {
		Warn("GetLocalByGW: parseIP %s failed", addr)
		return nil, nil
	}
	find := netlink.Route{LinkIndex: -1}
	for _, rte := range routes {
		if rte.Dst != nil && !rte.Dst.Contains(dest) {
			continue
		}
		if find.LinkIndex != -1 && find.Priority < rte.Priority {
			continue
		}
		find = rte
	}
	if find.LinkIndex != -1 {
		index := find.LinkIndex
		source := find.Gw
		if source == nil {
			source = find.Src
		}
		link, _ := netlink.LinkByIndex(index)
		address, _ := netlink.AddrList(link, netlink.FAMILY_V4)
		for _, ifAddr := range address {
			if ifAddr.Contains(source) {
				local = ifAddr.IP
			}
		}
	}
	Info("GetLocalByGw: find %s on %s", addr, local)
	return local, nil
}
