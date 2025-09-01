package schema

type BgpNeighbor struct {
	Address  string `json:"address"`
	RemoteAs int    `json:"remoteas"`
	State    string `json:"state"`
}

type Bgp struct {
	LocalAs   int           `json:"localas"`
	RouteId   string        `json:"routeid"`
	Neighbors []BgpNeighbor `json:"neighbors"`
}
