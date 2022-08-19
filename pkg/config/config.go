package config

type manager struct {
	Switch *Switch
}

var Manager = manager{
	Switch: &Switch{},
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
