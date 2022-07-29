package schema

type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address,omitempty"`
	Mac      string `json:"mac,omitempty"`
	Type     string `json:"type,omitempty"`
	Provider string `json:"provider"`
	Mtu      int    `json:"mtu,omitempty"`
}

type HwMacInfo struct {
	Uptime  int64  `json:"uptime"`
	Address string `json:"address"`
	Device  string `json:"device"`
}

type Bridge struct {
	Device
	Macs   []HwMacInfo `json:"macs"`
	Slaves []Device    `json:"slaves"`
	Stats  interface{} `json:"stats"`
}
