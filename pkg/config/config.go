package config

type manager struct {
	Switch *Switch
}

var Manager = manager{
	Switch: &Switch{},
}

func Reload() {
	Manager.Switch.Reload()
}
