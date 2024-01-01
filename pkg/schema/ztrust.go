package schema

type ZGuest struct {
	Network string `json:"network"`
	Name    string `json:"name"`
	Device  string `json:"device"`
	Address string `json:"Address"`
}

type KnockRule struct {
	Network string `json:"network"`
	Name    string `json:"name"`
	Dest    string `json:"destination"`
	Protocl string `json:"protocol"`
	Port    string `json:"port"`
	Age     int    `json:"age"`
}
