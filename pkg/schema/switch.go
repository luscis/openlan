package schema

type Switch struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Address string `json:"address"`
}

type LDAP struct {
	Server    string `json:"server"`
	BindDN    string `json:"bindDN"`
	BindPass  string `json:"bindPass"`
	BaseDN    string `json:"baseDN"`
	Attribute string `json:"attribute"`
	Filter    string `json:"filter"`
	EnableTls bool   `json:"enableTLS"`
	State     string `json:"state"`
}
