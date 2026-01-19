package network

import (
	"os/exec"
	"strings"
)

func LinkAdd(name string, opts ...string) ([]byte, error) {
	return nil, nil
}

func LinkSet(name string, opts ...string) ([]byte, error) {
	return nil, nil
}

func LinkDel(name string, opts ...string) ([]byte, error) {
	return nil, nil
}

func LinkUp(name string) ([]byte, error) {
	args := []string{
		"interface", "set", "interface",
		"name=" + name, "admin=ENABLED",
	}
	return exec.Command("netsh", args...).CombinedOutput()
}

func LinkDown(name string) ([]byte, error) {
	args := []string{
		"interface", "set", "interface",
		"name=" + name, "admin=DISABLED",
	}
	return exec.Command("netsh", args...).CombinedOutput()
}

func AddrAdd(name, addr string, opts ...string) ([]byte, error) {
	args := append([]string{
		"interface", "ipv4", "add", "address",
		"name=" + name, "address=" + addr, "store=active",
	}, opts...)
	return exec.Command("netsh", args...).CombinedOutput()
}

func AddrDel(name, addr string) ([]byte, error) {
	ipAddr := strings.SplitN(addr, "/", 1)[0]
	args := []string{
		"interface", "ipv4", "delete", "address",
		"name=" + name, "address=" + ipAddr, "store=active",
	}
	return exec.Command("netsh", args...).CombinedOutput()
}

func AddrShow(name string) []string {
	addrs := make([]string, 0, 4)
	args := []string{
		"interface", "ipv4", "show", "ipaddress",
		"interface=" + name, "level=verbose",
	}
	out, err := exec.Command("netsh", args...).Output()
	if err != nil {
		return nil
	}
	outArr := strings.Split(string(out), "\n")
	for _, addrStr := range outArr {
		addrArr := strings.SplitN(addrStr, " ", 3)
		if len(addrArr) != 3 {
			continue
		}
		if addrArr[0] == "Remote" && strings.Contains(addrArr[2], "Parameters") {
			addrs = append(addrs, addrArr[1])
		}
	}
	return addrs
}

func RouteAdd(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	args := []string{
		"interface", "ipv4", "add", "route",
		"prefix=" + prefix, "interface=" + name, "nexthop=" + nexthop,
		"store=active",
	}
	return exec.Command("netsh", args...).CombinedOutput()
}

func RouteDel(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	args := []string{
		"interface", "ipv4", "delete", "route",
		"prefix=" + prefix, "interface=" + name, "nexthop=" + nexthop,
		"store=active",
	}
	return exec.Command("netsh", args...).CombinedOutput()
}

func RouteShow(name string) []string {
	return nil
}

func GetDevInfo(name string) DeviceInfo {
	return DeviceInfo{}
}

func GetDevAddr(name string) []string {
	return nil
}

func RuleAdd(source string, lookup int, priority int) ([]byte, error) {
	return nil, nil
}

func RuleDel(source string, lookup int) ([]byte, error) {
	return nil, nil
}
