package config

type Bridge struct {
	Network  string `json:"network,omitempty"`
	Peer     string `json:"peer,omitempty"`
	Name     string `json:"name,omitempty"`
	IPMtu    int    `json:"mtu,omitempty"`
	Address  string `json:"address,omitempty"`
	Provider string `json:"provider,omitempty"`
	Stp      string `json:"stp,omitempty"`
	Delay    int    `json:"delay,omitempty"`
	Mss      int    `json:"tcpMss,omitempty"`
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
