package config

type manager struct {
	Switch *Switch
}

var Manager = manager{
	Switch: DefaultSwitch(),
}

func Reload() {
	Manager.Switch.Reload()
}
