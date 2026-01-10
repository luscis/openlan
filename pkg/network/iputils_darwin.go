package network

import (
	"os/exec"
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
	return nil, nil
}

func LinkDown(name string) ([]byte, error) {
	return nil, nil
}

func AddrAdd(name, addr string, opts ...string) ([]byte, error) {
	args := append([]string{
		name, addr,
	}, opts...)
	return exec.Command("ifconfig", args...).CombinedOutput()
}

func AddrDel(name, addr string) ([]byte, error) {
	args := []string{
		name, addr, "delete",
	}
	return exec.Command("ifconfig", args...).CombinedOutput()
}

func AddrShow(name string) []string {
	return nil
}

func RouteAdd(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	RouteDel("", prefix, nexthop)
	args := []string{"add", "-net", prefix}
	if name != "" {
		args = append(args, "-iface", name)
	}
	if nexthop != "" {
		args = append(args, nexthop)
	}
	args = append(args, opts...)
	return exec.Command("route", args...).CombinedOutput()
}

func RouteDel(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	args := []string{"delete", "-net", prefix}
	if name != "" {
		args = append(args, "-iface", name)
	}
	if nexthop != "" {
		args = append(args, nexthop)
	}
	args = append(args, opts...)
	return exec.Command("route", args...).CombinedOutput()
}

func RouteShow(name string) []string {
	return nil
}

func GetDevInfo(name string) DeviceInfo {
	return DeviceInfo{}
}

func GetDevAddr(name string) string {
	return ""
}

func RuleAdd(source string, lookup int, priority int) ([]byte, error) {
	return nil, nil
}

func RuleDel(source string, lookup int) ([]byte, error) {
	return nil, nil
}
