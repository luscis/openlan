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

func (t *RouterTunnel) Id() string {
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

type RouterSpecifies struct {
	Mss       int             `json:"tcpMss,omitempty" yaml:"tcpMss,omitempty"`
	Name      string          `json:"-" yaml:"-"`
	Private   []string        `json:"private,omitempty" yaml:"private,omitempty"`
	Loopback  string          `json:"loopback,omitempty" yaml:"loopback,omitempty"`
	Addresses []string        `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Tunnels   []*RouterTunnel `json:"tunnels,omitempty" yaml:"tunnels,omitempty"`
}

func (n *RouterSpecifies) Correct() {
	for _, t := range n.Tunnels {
		t.Correct()
	}
}

func (n *RouterSpecifies) FindTunnel(value *RouterTunnel) (*RouterTunnel, int) {
	for index, obj := range n.Tunnels {
		if obj.Id() == value.Id() {
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
