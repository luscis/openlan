package config

import "fmt"

type IPSecTunnel struct {
	Name      string `json:"-" yaml:"-"`
	Left      string `json:"local" yaml:"local"`
	LeftId    string `json:"localid,omitempty" yaml:"localid,omitempty"`
	LeftPort  int    `json:"localport,omitempty" yaml:"localport,omitempty"`
	Right     string `json:"remote" yaml:"remote"`
	RightId   string `json:"remoteid,omitempty" yaml:"remoteid,omitempty"`
	RightPort int    `json:"remoteport,omitempty" yaml:"remoteport,omitempty"`
	Transport string `json:"protocol" yaml:"protocol"`
	Secret    string `json:"secret" yaml:"secret"`
	State     string `json:"state" yaml:"state"`
}

func (s *IPSecTunnel) Correct() {
	if s.Left == "" {
		s.Left = "%defaultroute"
	}
	s.Name = s.Id()
	if s.RightId == "" {
		s.RightId = s.Right
	}
}

func (s *IPSecTunnel) Id() string {
	return fmt.Sprintf("%s-%s", s.Right, s.Transport)
}

type IPSecSpecifies struct {
	Name    string         `json:"-" yaml:"-"`
	Tunnels []*IPSecTunnel `json:"tunnels" yaml:"tunnels"`
}

func (s *IPSecSpecifies) Correct() {
	for _, t := range s.Tunnels {
		t.Correct()
	}
}

func (s *IPSecSpecifies) FindTunnel(value *IPSecTunnel) (*IPSecTunnel, int) {
	for index, obj := range s.Tunnels {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (s *IPSecSpecifies) AddTunnel(value *IPSecTunnel) bool {
	_, find := s.FindTunnel(value)
	if find == -1 {
		s.Tunnels = append(s.Tunnels, value)
	}
	return find == -1
}

func (s *IPSecSpecifies) DelTunnel(value *IPSecTunnel) (*IPSecTunnel, bool) {
	obj, find := s.FindTunnel(value)
	if find != -1 {
		s.Tunnels = append(s.Tunnels[:find], s.Tunnels[find+1:]...)
	}
	return obj, find != -1
}
