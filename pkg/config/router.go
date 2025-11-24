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
		t.Link = GenName("gre")
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
	Link      string          `json:"link,omitempty" yaml:"link,omitempty"`
	Subnets   []*Subnet       `json:"subnets,omitempty" yaml:"subnets,omitempty"`
	Loopback  string          `json:"loopback,omitempty" yaml:"loopback,omitempty"`
	Addresses []string        `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Tunnels   []*RouterTunnel `json:"tunnels,omitempty" yaml:"tunnels,omitempty"`
}

func (n *RouterSpecifies) Correct() {
	for _, t := range n.Tunnels {
		t.Correct()
	}
}
