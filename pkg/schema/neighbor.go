package schema

type Neighbor struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	HwAddr  string `json:"ethernet"`
	IpAddr  string `json:"address"`
	Client  string `json:"client"`
	Switch  string `json:"switch"`
	Network string `json:"network"`
	Device  string `json:"device"`
}
