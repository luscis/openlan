package schema

type Ceci struct {
	Name   string `json:"name"`
	Config any    `json:"config"`
}

type CeciProxy struct {
	Mode     string      `json:"mode"`
	Listen   string      `json:"listen"`
	Target   []string    `json:"target,omitempty"`
	Backends []ForwardTo `json:"backends,omitempty"`
	Cert     *Cert       `json:"cert,omitempty"`
	Status   string      `json:"status,omitempty"`
}

type ForwardTo struct {
	Server   string   `json:"server,omitempty"`
	Match    []string `json:"match,omitempty"`
	Protocol string   `json:"protocol,omitempty"`
	Insecure bool     `json:"insecure,omitempty"`
	Secret   string   `json:"secret,omitempty"`
	Nameto   string   `json:"nameto,omitempty"`
}

type Cert struct {
	CrtFile  string `json:"cert,omitempty"`
	KeyFile  string `json:"key,omitempty"`
	CaFile   string `json:"rootCa,omitempty"`
	CrtData  string `json:"certData,omitempty"`
	KeyData  string `json:"keyData,omitempty"`
	CaData   string `json:"rootCaData,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
}
