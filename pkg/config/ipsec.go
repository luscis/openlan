package config

import "fmt"

type IPSecTunnel struct {
	Name      string `json:"-"`
	Left      string `json:"local"`
	LeftId    string `json:"localid,omitempty"`
	LeftPort  int    `json:"localport,omitempty"`
	Right     string `json:"remote"`
	RightId   string `json:"remoteid,omitempty"`
	RightPort int    `json:"remoteport,omitempty"`
	Transport string `json:"protocol"`
	Secret    string `json:"secret"`
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
	Name    string         `json:"name"`
	Tunnels []*IPSecTunnel `json:"tunnels"`
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
