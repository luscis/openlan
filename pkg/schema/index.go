package schema

type Index struct {
	Version   Version          `json:"version"`
	Worker    Worker           `json:"worker"`
	Conntrack string           `json:"conntrack"`
	Access    []Access         `json:"access"`
	Devices   []Device         `json:"device"`
	Clients   []VPNClient      `json:"clients"`
	Outputs   []Output         `json:"output"`
	Routes    []KernelRoute    `json:"routes"`
	Neighbor  []KernelNeighbor `json:"neighbors"`
	UserLen   int              `json:"userLen"`
	CPUUsage  int              `json:"cpuUsage"`
	MemTotal  uint64           `json:"memTotal"`
	MemUsed   uint64           `json:"memUsed"`
	DiskTotal uint64           `json:"diskTotal"`
	DiskUsed  uint64           `json:"diskUsed"`
}

type Ctrl struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
