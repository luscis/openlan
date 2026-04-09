package schema

type Index struct {
	Version   Version          `json:"version"`
	Worker    Worker           `json:"worker"`
	Conntrack KernelConntrack  `json:"conntrack"`
	Access    []Access         `json:"access,omitempty"`
	Devices   []Device         `json:"device"`
	Clients   []VPNClient      `json:"clients,omitempty"`
	Outputs   []Output         `json:"output,omitempty"`
	Routes    []KernelRoute    `json:"routes,omitempty"`
	Neighbor  []KernelNeighbor `json:"neighbors,omitempty"`
	UserLen   int              `json:"userLen"`
	AccessLen int              `json:"accessLen"`
	ClientLen int              `json:"clientLen"`
	LinkLen   int              `json:"linkLen"`
	Usage     KernelUsage      `json:"usage"`
}

type Ctrl struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
