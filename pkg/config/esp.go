package config

import (
	"fmt"
	"github.com/luscis/openlan/pkg/libol"
	"net"
	"strconv"
	"strings"
)

var (
	EspAuth      = "8bc736635c0642aebc20ba5420c3e93a"
	EspCrypt     = "4ac161f6635843b8b02c60cc36822515"
	EspLocalUdp  = 4500
	EspRemoteUdp = 4500
)

func Addr2Cidr(addr string) string {
	if !strings.Contains(addr, "/") {
		return addr + "/32"
	}
	return addr
}

func SetLocalUdp(port string) {
	if udp, err := strconv.Atoi(port); err == nil {
		EspLocalUdp = udp
	}
}

type EspState struct {
	Local      string `json:"local,omitempty" yaml:"local,omitempty"`
	LocalIp    net.IP `json:"local_addr"  yaml:"localAddr"`
	Remote     string `json:"remote,omitempty" yaml:"remote,omitempty"`
	RemotePort int    `json:"remote_port" yaml:"remotePort"`
	RemoteIp   net.IP `json:"remote_addr"  yaml:"remoteAddr"`
	Encap      string `json:"encapsulation" yaml:"encapsulation"`
	Auth       string `json:"auth,omitempty" yaml:"auth,omitempty"`
	Crypt      string `json:"crypt,omitempty" yaml:"crypt,omitempty"`
}

func (s *EspState) Padding(value string, size int) string {
	return strings.Repeat(value, 64/len(value))[:size]
}

func (s *EspState) Correct(obj *EspState) {
	if obj != nil {
		if s.Local == "" {
			s.Local = obj.Local
		}
		if s.Auth == "" {
			s.Auth = obj.Auth
		}
		if s.Crypt == "" {
			s.Crypt = obj.Crypt
		}
		if s.RemotePort == 0 {
			s.RemotePort = obj.RemotePort
		}
	}
	if addr, _ := net.LookupIP(s.Local); len(addr) > 0 {
		s.LocalIp = addr[0]
	}
	if addr, _ := net.LookupIP(s.Remote); len(addr) > 0 {
		s.RemoteIp = addr[0]
	}
	if s.LocalIp == nil && s.RemoteIp != nil {
		addr, _ := libol.GetLocalByGw(s.RemoteIp.String())
		s.Local = addr.String()
		s.LocalIp = addr
	}
	if s.Crypt == "" {
		s.Crypt = s.Auth
	}
	if s.Auth == "" {
		s.Auth = EspAuth
	}
	if s.Crypt == "" {
		s.Crypt = EspCrypt
	}
	if s.Encap == "" {
		s.Encap = "udp"
	}
	if s.RemotePort == 0 {
		s.RemotePort = EspRemoteUdp
	}
	s.Auth = s.Padding(s.Auth, 32)
	s.Crypt = s.Padding(s.Crypt, 32)
}

type ESPPolicy struct {
	Source   string `json:"source,omitempty"`
	Dest     string `json:"destination,omitempty" yaml:"destination"`
	Priority int    `json:"priority"`
}

func (p *ESPPolicy) Correct() {
	if p.Source == "" {
		p.Source = "0.0.0.0/0"
	}
	p.Priority = 128 - libol.GetPrefixLen(p.Dest)
}

type ESPMember struct {
	Name     string       `json:"name"`
	Address  string       `json:"address,omitempty"`
	Peer     string       `json:"peer"`
	Spi      int          `json:"spi"`
	State    EspState     `json:"state"`
	Policies []*ESPPolicy `json:"policies" yaml:"policies"`
}

func (m *ESPMember) Correct(state *EspState) {
	if m.Name == "" {
		m.Name = fmt.Sprintf("spi:%d", m.Spi)
	} else if m.Spi == 0 {
		_, _ = fmt.Sscanf(m.Name, "spi:%d", &m.Spi)
	}
	if m.Address == "" || m.Peer == "" {
		return
	}
	m.Peer = Addr2Cidr(m.Peer)
	m.Address = Addr2Cidr(m.Address)
	ptr := &m.State
	ptr.Correct(state)
	if m.Policies == nil {
		m.Policies = make([]*ESPPolicy, 0, 2)
	}
	found := -1
	for index, pol := range m.Policies {
		pol.Correct()
		if pol.Dest != m.Peer {
			continue
		}
		found = index
	}
	if found < 0 {
		pol := &ESPPolicy{
			Dest: m.Peer,
		}
		pol.Correct()
		m.Policies = append(m.Policies, pol)
	}
}

func (m *ESPMember) AddPolicy(obj *ESPPolicy) {
	found := -1
	for index, po := range m.Policies {
		if po.Dest != obj.Dest {
			continue
		}
		found = index
		po.Source = obj.Source
		break
	}
	if found < 0 {
		obj.Correct()
		m.Policies = append(m.Policies, obj)
	}
}

func (m *ESPMember) RemovePolicy(dest string) bool {
	found := -1
	for index, po := range m.Policies {
		if po.Dest != dest {
			continue
		}
		found = index
		break
	}
	if found >= 0 {
		copy(m.Policies[found:], m.Policies[found+1:])
		m.Policies = m.Policies[:len(m.Policies)-1]
	}
	return found >= 0
}

type ESPSpecifies struct {
	Name    string       `json:"name"`
	Address string       `json:"address,omitempty"`
	State   EspState     `json:"state,omitempty" yaml:"state,omitempty"`
	Members []*ESPMember `json:"members"`
	Listen  string       `json:"listen,omitempty" yaml:"listen,omitempty"`
}

func (n *ESPSpecifies) Correct() {
	if n.Listen != "" {
		addr, port := libol.GetHostPort(n.Listen)
		if addr != "" {
			n.State.Local = addr
		}
		if port != "" {
			SetLocalUdp(port)
		}
	}
	for _, m := range n.Members {
		if m.Address == "" {
			m.Address = n.Address
		}
		m.Correct(&n.State)
	}
}

func (n *ESPSpecifies) GetMember(name string) *ESPMember {
	for _, mem := range n.Members {
		if mem.Name == name {
			return mem
		}
	}
	return nil
}

func (n *ESPSpecifies) HasRemote(name, addr string) bool {
	for _, mem := range n.Members {
		state := mem.State
		if state.Remote != name || state.RemoteIp.String() == addr {
			continue
		}
		return true
	}
	return false
}

func (n *ESPSpecifies) AddMember(obj *ESPMember) {
	found := -1
	for index, mem := range n.Members {
		if mem.Spi != obj.Spi && mem.Name != obj.Name {
			continue
		}
		found = index
		if len(obj.Policies) == 0 {
			obj.Policies = mem.Policies
		}
		n.Members[index] = obj
		break
	}
	if found < 0 {
		n.Members = append(n.Members, obj)
	}
}

func (n *ESPSpecifies) DelMember(name string) bool {
	found := -1
	for index, mem := range n.Members {
		if mem.Name != name {
			continue
		}
		found = index
		break
	}
	if found >= 0 {
		copy(n.Members[found:], n.Members[found+1:])
		n.Members = n.Members[:len(n.Members)-1]
	}
	return found >= 0
}
