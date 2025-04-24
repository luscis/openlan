package config

type LDAP struct {
	Server    string `json:"server" yaml:"server"`
	BindDN    string `json:"bindDN" yaml:"bindDN"`
	BindPass  string `json:"bindPass" yaml:"bindPass"`
	BaseDN    string `json:"baseDN" yaml:"baseDN"`
	Attribute string `json:"attribute" yaml:"attribute"`
	Filter    string `json:"filter" yaml:"filter"`
	Tls       bool   `json:"tLS" yaml:"tLS"`
}
