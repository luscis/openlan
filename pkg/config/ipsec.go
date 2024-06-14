package config

import "fmt"

type IPSecTunnel struct {
	Name      string `json:"-"`
	Left      string `json:"local"`
	LeftId    string `json:"localid"`
	LeftPort  string `json:"localport"`
	Right     string `json:"remote"`
	RightId   string `json:"remoteid"`
	RightPort string `json:"remoteport"`
	Transport string `json:"transport"`
	Secret    string `json:"secret"`
}

func (s *IPSecTunnel) Correct() {
	if s.Left == "" {
		s.Left = "%defaultroute"
	}
	s.Name = s.Id()
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
