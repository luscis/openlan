package config

type Output struct {
	Segment  int    `json:"segment"`
	Protocol string `json:"protocol,omitempty"` // gre, vxlan, etc.
	Remote   string `json:"remote"`
	DstPort  int    `json:"dstport,omitempty"`
	Link     string `json:"link,omitempty"` // link name
}
