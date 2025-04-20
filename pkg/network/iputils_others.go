//go:build !linux && !darwin && !windows

package network

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
	return nil, nil
}

func AddrDel(name, addr string) ([]byte, error) {
	return nil, nil
}

func AddrShow(name string) []string {
	return nil
}

func RouteAdd(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	return nil, nil
}

func RouteDel(name, prefix, nexthop string, opts ...string) ([]byte, error) {
	return nil, nil
}

func RouteShow(name string) []string {
	return nil
}
