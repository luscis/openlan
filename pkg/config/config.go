package config

type manager struct {
	Point  *Point
	Switch *Switch
	Proxy  *Proxy
}

var Manager = manager{
	Point:  &Point{},
	Switch: &Switch{},
	Proxy:  &Proxy{},
}

func GetNetwork(name string) *Network {
	for _, network := range Manager.Switch.Network {
		if network.Name == name {
			return network
		}
	}
	return nil
}

func Reload() {
	Manager.Switch.Reload()
}
