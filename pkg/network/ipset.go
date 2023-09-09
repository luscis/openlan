package network

import (
	"github.com/luscis/openlan/pkg/libol"
)

type IPSet struct {
	Name string
	Type string // hash:net, hash:ip
	Sudo bool
}

func NewIPSet(name, method string) *IPSet {
	return &IPSet{
		Name: "xtt_" + name,
		Type: method,
		Sudo: false,
	}
}

func (i *IPSet) exec(args ...string) (string, error) {
	bin := "/usr/sbin/ipset"
	if err := libol.FileExist(bin); err != nil {
		bin = "/sbin/ipset"
	}
	if i.Sudo {
		return libol.Sudo(bin, args...)
	} else {
		return libol.Exec(bin, args...)
	}
}

func (i *IPSet) Create() (string, error) {
	args := append([]string{"create", i.Name, i.Type, "-!"})
	return i.exec(args...)
}

func (i *IPSet) Clear() (string, error) {
	if out, err := i.Create(); err != nil {
		return out, err
	}
	if out, err := i.Flush(); err != nil {
		return out, err
	}
	return "", nil
}

func (i *IPSet) Destroy() (string, error) {
	args := append([]string{"destroy", i.Name})
	return i.exec(args...)
}

func (i *IPSet) Add(value string) (string, error) {
	args := append([]string{"add", i.Name, value})
	return i.exec(args...)
}

func (i *IPSet) Del(value string) (string, error) {
	args := append([]string{"del", i.Name, value})
	return i.exec(args...)
}

func (i *IPSet) Flush() (string, error) {
	args := append([]string{"flush", i.Name})
	return i.exec(args...)
}
