package config

type LDAP struct {
	Server    string `json:"server"`
	BindDN    string `json:"bindDN" yaml:"bindDN"`
	BindPass  string `json:"bindPass" yaml:"bindPass"`
	BaseDN    string `json:"baseDN" yaml:"baseDN"`
	Attribute string `json:"attribute"`
	Filter    string `json:"filter"`
	Tls       bool   `json:"tLS"`
}
