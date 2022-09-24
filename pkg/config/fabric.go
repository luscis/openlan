package config

type FabricSpecifies struct {
	Mss      int             `json:"tcpMss,omitempty"`
	Fragment bool            `json:"fragment"`
	Driver   string          `json:"driver,omitempty" yaml:"driver,omitempty"`
	Name     string          `json:"name,omitempty" yaml:"name,omitempty"`
	Tunnels  []*FabricTunnel `json:"tunnels" yaml:"tunnels"`
}

func (n *FabricSpecifies) Correct() {
	for _, tun := range n.Tunnels {
		tun.Correct()
		if tun.DstPort == 0 {
			if n.Driver == "stt" {
				tun.DstPort = 7471
			} else {
				tun.DstPort = 4789 // 8472
			}
		}
	}
}

func (n *FabricSpecifies) AddTunnel(obj *FabricTunnel) {
	found := -1
	for index, tun := range n.Tunnels {
		if tun.Remote != obj.Remote {
			continue
		}
		found = index
		n.Tunnels[index] = obj
		break
	}
	if found < 0 {
		n.Tunnels = append(n.Tunnels, obj)
	}
}

func (n *FabricSpecifies) DelTunnel(remote string) bool {
	found := -1
	for index, tun := range n.Tunnels {
		if tun.Remote != remote {
			continue
		}
		found = index
		break
	}
	if found >= 0 {
		copy(n.Tunnels[found:], n.Tunnels[found+1:])
		n.Tunnels = n.Tunnels[:len(n.Tunnels)-1]
	}
	return found >= 0
}

type FabricTunnel struct {
	DstPort uint32 `json:"dport" yaml:"destPort"`
	Remote  string `json:"remote"`
	Local   string `json:"local,omitempty" yaml:"local,omitempty"`
	Mode    string `json:"mode,omitempty" yaml:"mode,omitempty"`
}

func (c *FabricTunnel) Correct() {
}
