package config

import (
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
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
				Listen: "0.0.0.0:22",
				Target: []string{
					"1.1.1.1:23",
					"1.1.1.2:34",
				},
			},
		},
	}
	obj.Correct()
}

func TestHttpProxyCorrectDefaultPassword(t *testing.T) {
	obj := &HttpProxy{
		ConfDir: "/etc/openlan/http",
		Listen:  "0.0.0.0:1080",
	}
	obj.Correct()

	wantPassword := filepath.Join("/etc/openlan/http", "..", "switch", "password")
	if obj.Password != wantPassword {
		t.Fatalf("unexpected password path: got %q want %q", obj.Password, wantPassword)
	}
}

func TestHttpProxyMarshalOmitsNetwork(t *testing.T) {
	obj := &HttpProxy{
		Listen:    "0.0.0.0:1080",
		Network:   "guest",
		StatsFile: "/tmp/stats.json",
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "network:") {
		t.Fatalf("expected network to be omitted from yaml, got %s", data)
	}
	if strings.Contains(string(data), "statsfile:") {
		t.Fatalf("expected stats file to be omitted from yaml, got %s", data)
	}
}
