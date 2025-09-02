package schema

type BgpNeighbor struct {
	Address  string `json:"address"`
	RemoteAs int    `json:"remoteas"`
	State    string `json:"state"`
}

type Bgp struct {
	LocalAs   int           `json:"localas"`
	RouterId  string        `json:"routerid"`
	Neighbors []BgpNeighbor `json:"neighbors"`
}
