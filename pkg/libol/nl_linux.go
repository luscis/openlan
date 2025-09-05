package libol

import (
	"net"

	"github.com/vishvananda/netlink"
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

func ListRoutes() ([]Prefix, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	var data []Prefix
	for _, rte := range routes {
		link, err := netlink.LinkByIndex(rte.LinkIndex)
		if err != nil {
			continue
		}

		entry := Prefix{
			Protocol: rte.Protocol,
			Priority: rte.Priority,
			Link:     link.Attrs().Name,
		}

		if rte.Dst == nil {
			entry.Dst = "0.0.0.0/0"
		} else {
			entry.Dst = rte.Dst.String()
		}

		if len(rte.Gw) == 0 {
			entry.Gw = ""
		} else {
			entry.Gw = rte.Gw.String()
		}

		if len(rte.Src) == 0 {
			entry.Src = ""
		} else {
			entry.Src = rte.Src.String()
		}

		data = append(data, entry)
	}
	return data, nil
}
