package libol

import (
	"net"
	"os/exec"
	"runtime"
	"strings"
)

func IpLinkUp(name string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		args := []string{
			"link", "set", "dev", name, "up",
		}
		return exec.Command("ip", args...).CombinedOutput()
	case "windows":
		args := []string{
			"interface", "set", "interface",
			"name=" + name, "admin=ENABLED",
		}
		return exec.Command("netsh", args...).CombinedOutput()
	default:
		return nil, NewErr("IpLinkUp %s notSupport", runtime.GOOS)
	}
}

func IpLinkDown(name string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		args := []string{
			"link", "set", "dev", name, "down",
		}
		return exec.Command("ip", args...).CombinedOutput()
	case "windows":
		args := []string{
			"interface", "set", "interface",
			"name=" + name, "admin=DISABLED",
		}
		return exec.Command("netsh", args...).CombinedOutput()
	default:
		return nil, NewErr("IpLinkDown %s notSupport", runtime.GOOS)
	}
}

func IpAddrAdd(name, addr string, opts ...string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		args := append([]string{
			"address", "add", addr, "dev", name,
		}, opts...)
		return exec.Command("ip", args...).CombinedOutput()
	case "windows":
		args := append([]string{
			"interface", "ipv4", "add", "address",
			"name=" + name, "address=" + addr, "store=active",
		}, opts...)
		return exec.Command("netsh", args...).CombinedOutput()
	case "darwin":
		args := append([]string{
			name, addr,
		}, opts...)
		return exec.Command("ifconfig", args...).CombinedOutput()
	default:
		return nil, NewErr("IpAddrAdd %s notSupport", runtime.GOOS)
	}
}

func IpAddrDel(name, addr string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		args := []string{
			"address", "del", addr, "dev", name,
		}
		return exec.Command("ip", args...).CombinedOutput()
	case "windows":
		ipAddr := strings.SplitN(addr, "/", 1)[0]
		args := []string{
			"interface", "ipv4", "delete", "address",
			"name=" + name, "address=" + ipAddr, "store=active",
		}
		return exec.Command("netsh", args...).CombinedOutput()
	case "darwin":
		args := []string{
			name, addr, "delete",
		}
		return exec.Command("ifconfig", args...).CombinedOutput()
	default:
		return nil, NewErr("IpAddrDel %s notSupport", runtime.GOOS)
	}
}

func IpAddrShow(name string) []string {
	switch runtime.GOOS {
	case "windows":
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
	default:
		return nil
	}
}

func IpRouteAdd(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		args := []string{
			"route", "address", prefix, "via", nexthop,
		}
		return exec.Command("ip", args...).CombinedOutput()
	case "windows":
		args := []string{
			"interface", "ipv4", "add", "route",
			"prefix=" + prefix, "interface=" + name, "nexthop=" + nexthop,
			"store=active",
		}
		return exec.Command("netsh", args...).CombinedOutput()
	case "darwin":
		args := append([]string{
			"add", "-net", prefix})
		if name != "" {
			args = append(args, "-iface", name)
		}
		if nexthop != "" {
			args = append(args, nexthop)
		}
		args = append(args, opts...)
		return exec.Command("route", args...).CombinedOutput()
	default:
		return nil, NewErr("IpRouteAdd %s notSupport", runtime.GOOS)
	}
}

func IpRouteDel(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		args := []string{
			"route", "del", prefix, "via", nexthop,
		}
		return exec.Command("ip", args...).CombinedOutput()
	case "windows":
		args := []string{
			"interface", "ipv4", "delete", "route",
			"prefix=" + prefix, "interface=" + name, "nexthop=" + nexthop,
			"store=active",
		}
		return exec.Command("netsh", args...).CombinedOutput()
	case "darwin":
		args := append([]string{
			"delete", "-net", prefix})
		if name != "" {
			args = append(args, "-iface", name)
		}
		if nexthop != "" {
			args = append(args, nexthop)
		}
		args = append(args, opts...)
		return exec.Command("route", args...).CombinedOutput()
	default:
		return nil, NewErr("IpRouteDel %s notSupport", runtime.GOOS)
	}
}

func IpRouteShow(name string) []string {
	switch runtime.GOOS {
	default:
		return nil
	}
}

func IpMetricSet(name, metric string, opts ...string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		return nil, nil
	case "windows":
		args := append([]string{
			"interface", "ipv4", "set", "interface",
			"interface=" + name, "metric=" + metric,
		}, opts...)
		return exec.Command("netsh", args...).CombinedOutput()
	case "darwin":
		//TODO
		return nil, nil
	default:
		return nil, NewErr("IpAddrAdd %s notSupport", runtime.GOOS)
	}
}

func LookupIP(name string) string {
	if addr, _ := net.LookupIP(name); len(addr) > 0 {
		return addr[0].String()
	}
	return ""
}
