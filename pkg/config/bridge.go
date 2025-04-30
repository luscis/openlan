package config

type Bridge struct {
	Network  string `json:"-" yaml:"-"`
	Share    string `json:"share,omitempty" yaml:"share,omitempty"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	IPMtu    int    `json:"mtu,omitempty" yaml:"mtu,omitempty"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Provider string `json:"-" yaml:"-"`
	Stp      string `json:"stp,omitempty" yaml:"stp,omitempty"`
	Delay    int    `json:"delay,omitempty" yaml:"delay,omitempty"`
	Mss      int    `json:"tcpMss,omitempty" yaml:"tcpMss,omitempty"`
}

func (br *Bridge) Correct() {
	if br.Name == "" {
		if len(br.Network) > 12 {
			br.Name = "br-" + br.Network[:12]
		} else {
			br.Name = "br-" + br.Network
		}
	}
	if br.Provider == "" {
		br.Provider = "linux"
	}
	if br.IPMtu == 0 {
		br.IPMtu = 1500
	}
	if br.Delay == 0 {
		br.Delay = 2
	}
	if br.Stp == "" {
		br.Stp = "enable"
	}
}
