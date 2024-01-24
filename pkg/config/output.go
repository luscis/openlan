package config

type Output struct {
	Segment  int    `json:"segment"`
	Protocol string `json:"protocol"` // gre, vxlan, etc.
	Remote   string `json:"remote"`
	DstPort  int    `json:"dstport"`
	Link     string `json:"link"` // link name
}
