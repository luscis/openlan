package config

type IPSecTunnel struct {
	Left      string `json:"local"`
	LeftId    string `json:"localid"`
	LeftPort  string `json:"localport"`
	Right     string `json:"remote"`
	RightId   string `json:"remoteid"`
	RightPort string `json:"remoteport"`
	Transport string `json:"transport"`
}

func (s *IPSecTunnel) Correct() {
	if s.Left == "" {
		s.Left = "%defaultroute"
	}
}

type IPSecSpecifies struct {
	Name    string        `json:"name"`
	Tunnels []IPSecTunnel `json:"tunnels"`
}

func (s *IPSecSpecifies) Correct() {
	for _, t := range s.Tunnels {
		t.Correct()
	}
}
