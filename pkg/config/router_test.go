package config

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestRouterSpecifiesAddressDevice(t *testing.T) {
	spec := &RouterSpecifies{}
	if ok := spec.AddAddress(&RouterAddress{Device: "eth0", Address: "192.0.2.1/24"}); !ok {
		t.Fatal("expected address to be added")
	}
	if ok := spec.AddAddress(&RouterAddress{Device: "eth0", Address: "192.0.2.1/24"}); ok {
		t.Fatal("expected duplicate address on same device to be ignored")
	}
	if ok := spec.AddAddress(&RouterAddress{Device: "eth1", Address: "192.0.2.1/24"}); ok {
		t.Fatal("expected duplicate address on different device to be ignored")
	}
	if ok := spec.AddAddress(&RouterAddress{Device: "eth1", Address: "198.51.100.1/24"}); !ok {
		t.Fatal("expected different address to be added")
	}

	old, ok := spec.DelAddress(&RouterAddress{Address: "192.0.2.1/24"})
	if !ok || old.Device != "eth0" {
		t.Fatalf("expected eth0 address to be removed, got %#v %v", old, ok)
	}
	if len(spec.Addresses) != 1 || spec.Addresses[0].Address != "198.51.100.1/24" {
		t.Fatalf("unexpected remaining addresses: %#v", spec.Addresses)
	}
}

func TestRouterAddressUnmarshalLegacyString(t *testing.T) {
	var addr RouterAddress
	if err := json.Unmarshal([]byte(`"198.51.100.1/32"`), &addr); err != nil {
		t.Fatal(err)
	}
	if addr.Device != "lo" || addr.Address != "198.51.100.1/32" {
		t.Fatalf("unexpected address: %#v", addr)
	}
}

func TestRouterSpecifiesUnmarshalLegacyStringAddresses(t *testing.T) {
	var spec RouterSpecifies
	if err := yaml.Unmarshal([]byte("addresses:\n- 198.51.100.1/32\n"), &spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Addresses) != 1 || spec.Addresses[0].Device != "lo" || spec.Addresses[0].Address != "198.51.100.1/32" {
		t.Fatalf("unexpected addresses: %#v", spec.Addresses)
	}
}
