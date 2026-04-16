package schema

type Ceci struct {
	Name   string `json:"name"`
	Config any    `json:"config"`
}

type CeciProxy struct {
	Mode     string      `json:"mode"`
	Listen   string      `json:"listen"`
	Network  string      `json:"network,omitempty"`
	Target   []string    `json:"target,omitempty"`
	Backends []ForwardTo `json:"backends,omitempty"`
	Cert     *Cert       `json:"cert,omitempty"`
	Stats    *CeciStats  `json:"stats,omitempty"`
	Status   string      `json:"status,omitempty"`
}

type CeciStats struct {
	StartAt string `json:"startAt,omitempty"`
	Total   int    `json:"total,omitempty"`
	Bytes   int64  `json:"bytes,omitempty"`
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
