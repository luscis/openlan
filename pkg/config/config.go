package config

var switcher *Switch

func Reload() {
	switcher.Reload()
}

func GetAcl(name string) *ACL {
	return switcher.GetACL(name)
}

func Update(obj *Switch) {
	switcher = obj
}

func Get() *Switch {
	return switcher
}
