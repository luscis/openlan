package config

import "fmt"

type Linker interface {
	Start() error
	Stop() error
}

type Output struct {
	Segment  int    `json:"segment" yaml:"segment"`
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"` // gre, vxlan, tcp/tls/wss etc.
	Remote   string `json:"remote" yaml:"remote"`
	Fallback string `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	DstPort  int    `json:"dstport,omitempty" yaml:"dstport,omitempty"`
	Link     string `json:"link,omitempty" yaml:"link,omitempty"` // link name
	Secret   string `json:"secret,omitempty" yaml:"secret,omitempty"`
	Crypt    string `json:"crypt,omitempty" yaml:"crypt,omitempty"`
	Linker   Linker `json:"-" yaml:"-"`
}

func (o *Output) Id() string {
	return fmt.Sprintf("%s-%s-%d", o.Protocol, o.Remote, o.Segment)
}

func (o *Output) Correct() {
	switch o.Protocol {
	case "gre":
		o.Link = fmt.Sprintf("%s%d", "gei", o.Segment)
	case "vxlan":
		o.Link = fmt.Sprintf("%s%d", "xei", o.Segment)
	case "tcp", "tls", "wss":
		o.Link = o.Remote
	default:
		if o.Segment > 0 {
			o.Link = fmt.Sprintf("%s.%d", o.Remote, o.Segment)
		} else {
			o.Link = o.Remote
		}
	}
}
