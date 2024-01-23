package config

type Output struct {
	Segment  int    `json:"segment"`
	Protocol string `json:"protocol"` // gre, vxlan, etc.
	Remote   string `json:"remote"`
	Link     string `json:"link"` // link name
}
