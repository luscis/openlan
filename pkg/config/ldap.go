package config

type LDAP struct {
	Server    string `json:"server"`
	BindDN    string `json:"bindDN"`
	BindPass  string `json:"bindPass"`
	BaseDN    string `json:"baseDN"`
	Attribute string `json:"attribute"`
	Filter    string `json:"filter"`
	EnableTls bool   `json:"enableTLS"`
}
