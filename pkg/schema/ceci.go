package schema

type Ceci struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type CeciTcp struct {
	Mode   string   `json:"mode"`
	Listen string   `json:"listen"`
	Target []string `json:"target,omitempty"`
}
