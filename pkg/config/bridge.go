package config

type Bridge struct {
	Network  string `json:"-" yaml:"-"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	IPMtu    int    `json:"-" yaml:"-"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Provider string `json:"-" yaml:"-"`
	Stp      string `json:"-" yaml:"-"`
	Delay    int    `json:"-" yaml:"-"`
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
