package config

var switcher *Switch

func Reload() {
	switcher.Reload()
}

func GetAcl(name string) *ACL {
	return switcher.GetACL(name)
}

func GetQos(name string) *Qos {
	return switcher.GetQos(name)
}

func Update(obj *Switch) {
	switcher = obj
}

func Get() *Switch {
	return switcher
}

func GetNetwork(name string) *Network {
	return switcher.GetNetwork(name)
}
