package network

import (
	"os/exec"
	"strconv"
	"strings"

	nl "github.com/vishvananda/netlink"
)

func LinkAdd(name string, opts ...string) ([]byte, error) {
	args := append([]string{
		"link", "add", name,
	}, opts...)
	return exec.Command("ip", args...).CombinedOutput()
}

func LinkSet(name string, opts ...string) ([]byte, error) {
	args := append([]string{
		"link", "set", name,
	}, opts...)
	return exec.Command("ip", args...).CombinedOutput()
}

func LinkDel(name string, opts ...string) ([]byte, error) {
	args := append([]string{
		"link", "del", name,
	}, opts...)
	return exec.Command("ip", args...).CombinedOutput()
}

func LinkUp(name string) ([]byte, error) {
	args := []string{
		"link", "set", "dev", name, "up",
	}
	return exec.Command("ip", args...).CombinedOutput()
}

func LinkDown(name string) ([]byte, error) {
	args := []string{
		"link", "set", "dev", name, "down",
	}
	return exec.Command("ip", args...).CombinedOutput()
}

func AddrAdd(name, addr string, opts ...string) ([]byte, error) {
	args := append([]string{
		"address", "add", addr, "dev", name,
	}, opts...)
	return exec.Command("ip", args...).CombinedOutput()
}

func AddrDel(name, addr string) ([]byte, error) {
	args := []string{
		"address", "del", addr, "dev", name,
	}
	return exec.Command("ip", args...).CombinedOutput()
}

func AddrShow(name string) []string {
	return nil
}

func RouteAdd(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	args := []string{
		"route", "replace", prefix, "via", nexthop,
	}
	args = append(args, opts...)
	return exec.Command("ip", args...).CombinedOutput()
}

func RouteDel(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	args := []string{
		"route", "del", prefix, "via", nexthop,
	}
	return exec.Command("ip", args...).CombinedOutput()
}

func RouteShow(name string) []string {
	return nil
}

func GetDevInfo(name string) DeviceInfo {
	if link, err := nl.LinkByName(name); err == nil {
		attr := link.Attrs()
		state := "down"
		if strings.Contains(attr.Flags.String(), "up") {
			state = "up"
		}
		return DeviceInfo{
			State: state,
			Drop:  attr.Statistics.RxDropped,
			Recv:  attr.Statistics.RxBytes,
			Send:  attr.Statistics.TxBytes,
			Mac:   attr.HardwareAddr.String(),
			Mtu:   attr.MTU,
		}
	}
	return DeviceInfo{}
}

func GetDevAddr(name string) []string {
	addrs := []string{}
	if link, err := nl.LinkByName(name); err == nil {
		address, _ := nl.AddrList(link, nl.FAMILY_V4)
		for _, addr := range address {
			addrs = append(addrs, addr.IPNet.String())
		}
	}
	return addrs
}

func RuleAdd(source string, lookup int, priority int) ([]byte, error) {
	args := []string{
		"rule", "add",
		"from", source,
		"lookup", strconv.Itoa(lookup),
	}
	if priority > 0 {
		args = append(args, "priority", strconv.Itoa(priority))
	}
	return exec.Command("ip", args...).CombinedOutput()
}

func RuleDel(source string, lookup int) ([]byte, error) {
	args := []string{
		"rule", "del",
		"from", source,
		"lookup", strconv.Itoa(lookup),
	}
	return exec.Command("ip", args...).CombinedOutput()
}
