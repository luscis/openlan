package schema

type BgpNeighbor struct {
	Address  string   `json:"address"`
	RemoteAs int      `json:"remoteas"`
	State    string   `json:"state,omitempty" yaml:"state,omitempty"`
	Advertis []string `json:"advertis"`
	Receives []string `json:"receives"`
}

type Bgp struct {
	LocalAs   int           `json:"localas"`
	RouterId  string        `json:"routerid"`
	Neighbors []BgpNeighbor `json:"neighbors"`
}

type BgpPrefix struct {
	Prefix   string `json:"prefix"`
	Neighbor string `json:"neighbor"`
}
