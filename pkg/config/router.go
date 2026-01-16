package config

import (
	"fmt"
	"strings"
)

type RouterTunnel struct {
	Link     string `json:"link,omitempty" yaml:"link,omitempty"`
	Remote   string `json:"remote,omitempty" yaml:"remote,omitempty"`
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
}

func (t *RouterTunnel) ID() string {
	return fmt.Sprintf("%s-%s", t.Protocol, t.Remote)
}

func (t *RouterTunnel) Correct() {
	if t.Protocol == "" {
		t.Protocol = "gre"
	}
	switch t.Protocol {
	case "gre":
		t.Link = GenName("ige")
	case "ipip":
		t.Link = GenName("ipi")
	}

	if t.Address != "" && !strings.Contains(t.Address, "/") {
		t.Address = t.Address + "/30"
	}
}

type RouterRedirect struct {
	Source   string `json:"source,omitempty" yaml:"source,omitempty"`
	NextHop  string `json:"nexthop,omitempty" yaml:"nexthop,omitempty"`
	Table    int    `json:"table,omitempty" yaml:"table,omitempty"` // 1-250
	Priority int    `json:"priority,omitempty" yaml:"priority,omitempty"`
}

func (r *RouterRedirect) Correct() {
	if r.Priority == 0 {
		r.Priority = r.Table + 100
	}
}

func (r *RouterRedirect) ID() string {
	return fmt.Sprintf("%s", r.Source)
}
func (r *RouterRedirect) Rule() string {
	return fmt.Sprintf("from %s lookup %d", r.Source, r.Table)
}

func (r *RouterRedirect) Route() string {
	return fmt.Sprintf("default via %s table %d", r.NextHop, r.Table)
}

type RouterInterface struct {
	Device  string `json:"device,omitempty" yaml:"device,omitempty"`
	VLAN    int    `json:"vlan,omitempty" yaml:"vlan,omitempty"`
	Address string `json:"address,omitempty" yaml:"address,omitempty"`
}

func (i *RouterInterface) ID() string {
	if i.VLAN == 0 {
		return i.Device
	}
	return fmt.Sprintf("%s.%d", i.Device, i.VLAN)
}

type RouterSpecifies struct {
	Mss        int                `json:"tcpMss,omitempty" yaml:"tcpMss,omitempty"`
	Name       string             `json:"-" yaml:"-"`
	Private    []string           `json:"private,omitempty" yaml:"private,omitempty"`
	Loopback   string             `json:"loopback,omitempty" yaml:"loopback,omitempty"`
	Addresses  []string           `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Tunnels    []*RouterTunnel    `json:"tunnels,omitempty" yaml:"tunnels,omitempty"`
	Interfaces []*RouterInterface `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	Redirect   []*RouterRedirect  `json:"redirect,omitempty" yaml:"redirect,omitempty"`
}

func (n *RouterSpecifies) Correct() {
	for _, t := range n.Tunnels {
		t.Correct()
	}
	for _, t := range n.Redirect {
		t.Correct()
	}
}

func (n *RouterSpecifies) FindTunnel(value *RouterTunnel) (*RouterTunnel, int) {
	for index, obj := range n.Tunnels {
		if obj.ID() == value.ID() {
			return obj, index
		}
	}
	return nil, -1
}

func (n *RouterSpecifies) AddTunnel(value *RouterTunnel) bool {
	_, index := n.FindTunnel(value)
	if index == -1 {
		n.Tunnels = append(n.Tunnels, value)
		return true
	}
	return false
}

func (n *RouterSpecifies) DelTunnel(value *RouterTunnel) (*RouterTunnel, bool) {
	older, index := n.FindTunnel(value)
	if index != -1 {
		n.Tunnels = append(n.Tunnels[:index], n.Tunnels[index+1:]...)
		return older, true
	}
	return older, false
}

func (n *RouterSpecifies) FindPrivate(value string) (string, int) {
	for index, obj := range n.Private {
		if value == obj {
			return obj, index
		}
	}
	return "", -1
}

func (n *RouterSpecifies) AddPrivate(value string) bool {
	_, index := n.FindPrivate(value)
	if index == -1 {
		n.Private = append(n.Private, value)
		return true
	}
	return false
}

func (n *RouterSpecifies) DelPrivate(value string) (string, bool) {
	older, index := n.FindPrivate(value)
	if index != -1 {
		n.Private = append(n.Private[:index], n.Private[index+1:]...)
		return older, true
	}
	return older, false
}

func (n *RouterSpecifies) FindInterface(value *RouterInterface) (*RouterInterface, int) {
	for index, obj := range n.Interfaces {
		if value.ID() == obj.ID() {
			return obj, index
		}
	}
	return nil, -1
}

func (n *RouterSpecifies) AddInterface(value *RouterInterface) bool {
	_, index := n.FindInterface(value)
	if index == -1 {
		n.Interfaces = append(n.Interfaces, value)
		return true
	}
	return false
}

func (n *RouterSpecifies) DelInterface(value *RouterInterface) (*RouterInterface, bool) {
	older, index := n.FindInterface(value)
	if index != -1 {
		n.Interfaces = append(n.Interfaces[:index], n.Interfaces[index+1:]...)
		return older, true
	}
	return older, false
}

func (n *RouterSpecifies) FindRedirect(value *RouterRedirect) (*RouterRedirect, int) {
	for index, obj := range n.Redirect {
		if value.ID() == obj.ID() {
			return obj, index
		}
	}
	return nil, -1
}

func (n *RouterSpecifies) AddRedirect(value *RouterRedirect) bool {
	_, index := n.FindRedirect(value)
	if index == -1 {
		n.Redirect = append(n.Redirect, value)
		return true
	}
	return false
}

func (n *RouterSpecifies) DelRedirect(value *RouterRedirect) (*RouterRedirect, bool) {
	older, index := n.FindRedirect(value)
	if index != -1 {
		n.Redirect = append(n.Redirect[:index], n.Redirect[index+1:]...)
		return older, true
	}
	return older, false
}
