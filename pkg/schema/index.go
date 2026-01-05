package schema

type Index struct {
	Version   Version     `json:"version"`
	Worker    Worker      `json:"worker"`
	Conntrack string      `json:"conntrack"`
	Access    []Access    `json:"access"`
	Neighbors []Neighbor  `json:"neighbors"`
	OnLines   []OnLine    `json:"online"`
	Network   []Network   `json:"network"`
	Clients   []VPNClient `json:"clients"`
	Outputs   []Output    `json:"output"`
}

type Ctrl struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
