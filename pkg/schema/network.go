package schema

type Lease struct {
	Address string `json:"address"`
	Alias   string `json:"alias"`
	Client  string `json:"client"`
	Type    string `json:"type"`
	Network string `json:"network"`
}

type PrefixRoute struct {
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop,omitempty"`
	FindHop   string      `json:"findhop,omitempty"`
	Metric    int         `json:"metric"`
	Link      string      `json:"link,omitempty"`
	Table     int         `json:"table,omitempty"`
	Source    string      `json:"source,omitempty"`
	Protocol  string      `json:"protocol,omitempty"`
	MultiPath []MultiPath `json:"multipath,omitempty"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Weight  int    `json:"weight"`
}

type Subnet struct {
	IfAddr  string `json:"address,omitempty"`
	IpStart string `json:"startAt,omitempty"`
	IpEnd   string `json:"endAt,omitempty"`
	Netmask string `json:"netmask"`
}

type Network struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type FindHop struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	Check     string `json:"check"`
	NextHop   string `json:"nexthop"`
	Available string `json:"available"`
}
