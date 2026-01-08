package libol

import (
	"fmt"
	"net"

	nl "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func GetLocalByGw(addr string) (net.IP, error) {
	local := net.IP{}
	routes, err := nl.RouteList(nil, nl.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	dest := net.ParseIP(addr)
	if dest == nil {
		Warn("GetLocalByGW: parseIP %s failed", addr)
		return nil, nil
	}
	find := nl.Route{LinkIndex: -1}
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
		link, _ := nl.LinkByIndex(index)
		address, _ := nl.AddrList(link, nl.FAMILY_V4)
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
	var items []Prefix

	values, err := nl.RouteList(nil, nl.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		entry := Prefix{
			Protocol: RouteProtocol(value.Protocol),
			Priority: value.Priority,
		}
		link, err := nl.LinkByIndex(value.LinkIndex)
		if err == nil {
			entry.Link = link.Attrs().Name
		}
		if value.Dst == nil {
			entry.Dst = "0.0.0.0/0"
		} else {
			entry.Dst = value.Dst.String()
		}
		if len(value.Gw) == 0 {
			entry.Gw = "0.0.0.0"
		} else {
			entry.Gw = value.Gw.String()
		}
		if len(value.Src) == 0 {
			entry.Src = "0.0.0.0"
		} else {
			entry.Src = value.Src.String()
		}
		items = append(items, entry)
	}
	return items, nil
}

func StateCode(code int) string {
	switch code {
	case 0x00:
		return "NONE"
	case 0x01:
		return "INCOMPLETE"
	case 0x02:
		return "REACHABLE"
	case 0x04:
		return "STALE"
	case 0x08:
		return "DELAY"
	case 0x10:
		return "PROBE"
	case 0x20:
		return "FAILED"
	case 0x40:
		return "NOARP"
	case 0x80:
		return "PERMANENT"
	default:
		return fmt.Sprintf("%d", code)
	}
}

func ListNeighbrs() ([]Neighbor, error) {
	var items []Neighbor

	values, err := nl.NeighList(0, nl.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		entry := Neighbor{
			Address: value.IP.String(),
			HwAddr:  value.HardwareAddr.String(),
			State:   StateCode(value.State),
		}
		link, err := nl.LinkByIndex(value.LinkIndex)
		if err == nil {
			entry.Link = link.Attrs().Name
		}
		items = append(items, entry)
	}
	return items, nil
}

func ListConnStats() ConnStats {
	sts := ConnStats{}
	values, err := nl.ConntrackTableList(nl.ConntrackTable, nl.FAMILY_V4)
	if err != nil {
		return sts
	}
	for _, value := range values {
		sts.Total += 1
		switch value.Forward.Protocol {
		case 6:
			sts.TCP += 1
		case 17:
			sts.UDP += 1
		case 1:
			sts.ICMP += 1
		}
	}
	return sts
}

func ListPhyLinks() []Device {
	var dev []Device
	values, err := nl.LinkList()
	if err != nil {
		return dev
	}
	for _, value := range values {
		t := value.Type()
		if t == "device" || t == "vlan" || t == "ipip" || t == "gre" {
			attr := value.Attrs()
			state := "down"
			if attr.Flags&unix.IFF_UP != 0 {
				state = "up"
			}
			dev = append(dev, Device{
				Name:  attr.Name,
				Mtu:   attr.MTU,
				Mac:   attr.HardwareAddr.String(),
				State: state,
				Drop:  attr.Statistics.RxDropped,
				Recv:  attr.Statistics.RxBytes,
				Send:  attr.Statistics.TxPackets,
			})
		}
	}
	return dev
}
