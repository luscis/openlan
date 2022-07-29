package config

type Password struct {
	Network  string `json:"network,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
}
