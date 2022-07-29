package config

type LDAP struct {
	Server    string `json:"server"`
	BindDN    string `json:"bindDN"`
	Password  string `json:"password"`
	BaseDN    string `json:"baseDN"`
	Attribute string `json:"attribute"`
	Filter    string `json:"filter"`
	EnableTls bool   `json:"enableTLS"`
}
