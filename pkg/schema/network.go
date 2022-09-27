package schema

type Lease struct {
	Address string `json:"address"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Client  string `json:"client"`
	Type    string `json:"type"`
	Network string `json:"network"`
}

type PrefixRoute struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"nexthop"`
	Metric  int    `json:"metric"`
	Mode    string `json:"mode"`
}

type Subnet struct {
	IfAddr  string `json:"address,omitempty"`
	IpStart string `json:"startAt,omitempty"`
	IpEnd   string `json:"endAt,omitempty"`
	Netmask string `json:"netmask"`
}

type Network struct {
	Name   string        `json:"name"`
	Subnet Subnet        `json:"subnet"`
	Routes []PrefixRoute `json:"routes"`
}
