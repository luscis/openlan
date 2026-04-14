package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestCeciProxyMarshalOmitsNetwork(t *testing.T) {
	obj := &CeciProxy{
		Mode:    "http",
		Listen:  "127.0.0.1:8080",
		Network: "guest",
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "network:") {
		t.Fatalf("expected network to be omitted from yaml, got %s", data)
	}
}
