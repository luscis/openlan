package network

import "github.com/luscis/openlan/pkg/libol"

const (
	IPSetBin = "/usr/sbin/ipset"
)

type IPSet struct {
	Name string
	Type string // hash:net, hash:ip
}

func NewIPSet(name, method string) *IPSet {
	return &IPSet{
		Name: name,
		Type: method,
	}
}

func (i *IPSet) Create() (string, error) {
	args := append([]string{"create", i.Name, i.Type, "-!"})
	return libol.Sudo(IPSetBin, args...)
}

func (i *IPSet) Destroy() (string, error) {
	args := append([]string{"destroy", i.Name})
	return libol.Sudo(IPSetBin, args...)
}

func (i *IPSet) Add(value string) (string, error) {
	args := append([]string{"add", i.Name, value})
	return libol.Sudo(IPSetBin, args...)
}

func (i *IPSet) Del(value string) (string, error) {
	args := append([]string{"del", i.Name, value})
	return libol.Sudo(IPSetBin, args...)
}

func (i *IPSet) Flush() (string, error) {
	args := append([]string{"flush", i.Name})
	return libol.Sudo(IPSetBin, args...)
}
