package config

import (
	"testing"
)

func TestNewProxy(t *testing.T) {
	obj := &Proxy{
	Http: []*HttpProxy{
		{
			Listen: "0.0.0.0:80",
			Cert: &Cert{
				Dir: "/var/run",
			},
		},
	},
	Tcp: []*TcpProxy{
		{
			Listen:"0.0.0.0:22",
			Target: []string{
				"1.1.1.1:23",
				"1.1.1.2:34",
			},
		},
	},
	}
	obj.Correct()
}