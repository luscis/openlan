package libol

import (
	"fmt"
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

const (
	RTPROT_BGP      = 0xba
	RTPROT_BIRD     = 0xc
	RTPROT_KERNEL   = 0x2
	RTPROT_BOOT     = 0x3
	RTPROT_DHCP     = 0x10
	RTPROT_ISIS     = 0xbb
	RTPROT_OSPF     = 0xbc
	RTPROT_REDIRECT = 0x1
	RTPROT_RIP      = 0xbd
	RTPROT_STATIC   = 0x4
	RTPROT_UNSPEC   = 0x0
	RTPROT_ZEBRA    = 0xb
)

func RouteProtocol(code int) string {
	switch code {
	case RTPROT_BGP:
		return "bgp"
	case RTPROT_KERNEL:
		return "kernel"
	case RTPROT_BOOT:
		return "boot"
	case RTPROT_DHCP:
		return "dhcp"
	case RTPROT_ISIS:
		return "isis"
	case RTPROT_OSPF:
		return "ospf"
	case RTPROT_RIP:
		return "rip"
	case RTPROT_STATIC:
		return "static"
	default:
		return fmt.Sprintf("%d", code)
	}
}

func ListRoutes() ([]Prefix, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	var data []Prefix
	for _, rte := range routes {
		entry := Prefix{
			Protocol: RouteProtocol(rte.Protocol),
			Priority: rte.Priority,
		}
		link, err := netlink.LinkByIndex(rte.LinkIndex)
		if err == nil {
			entry.Link = link.Attrs().Name
		}
		if rte.Dst == nil {
			entry.Dst = "0.0.0.0/0"
		} else {
			entry.Dst = rte.Dst.String()
		}

		if len(rte.Gw) == 0 {
			entry.Gw = "0.0.0.0"
		} else {
			entry.Gw = rte.Gw.String()
		}

		if len(rte.Src) == 0 {
			entry.Src = "0.0.0.0"
		} else {
			entry.Src = rte.Src.String()
		}

		data = append(data, entry)
	}
	return data, nil
}
