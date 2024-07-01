package config

import "fmt"

type Linker interface {
	Start() error
	Stop() error
}
type Output struct {
	Segment  int    `json:"segment"`
	Protocol string `json:"protocol,omitempty"` // gre, vxlan, tcp/tls/wss etc.
	Remote   string `json:"remote"`
	DstPort  int    `json:"dstport,omitempty"`
	Link     string `json:"link,omitempty"` // link name
	Secret   string `json:"secret,omitempty"`
	Crypt    string `json:"crypt,omitempty"`
	Linker   Linker `json:"-"`
}

func (o *Output) Id() string {
	return fmt.Sprintf("%s-%s-%d", o.Protocol, o.Remote, o.Segment)
}

func (o *Output) GenName() {
	if o.Link == "" {
		if o.Protocol == "gre" {
			o.Link = fmt.Sprintf("%s%d", "gre", o.Segment)
		} else if o.Protocol == "vxlan" {
			o.Link = fmt.Sprintf("%s%d", "vxlan", o.Segment)
		} else if o.Protocol == "tcp" || o.Protocol == "tls" ||
			o.Protocol == "wss" {
			o.Link = o.Remote
		} else if o.Segment > 0 {
			o.Link = fmt.Sprintf("%s.%d", o.Remote, o.Segment)
		} else {
			o.Link = o.Remote
		}
	}
}
