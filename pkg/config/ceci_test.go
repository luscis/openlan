package config

import (
	"os"
	"path/filepath"
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

func TestCertCorrectKeepsContentOnly(t *testing.T) {
	cert := &Cert{
		CrtData: "crt-data",
		KeyData: "key-data",
		CaData:  "ca-data",
	}
	cert.Correct()
	if cert.CrtFile != "" || cert.KeyFile != "" || cert.CaFile != "" {
		t.Fatalf("expected content-only cert to keep file paths empty, got %+v", cert)
	}
}

func TestCertLoadDataFromFiles(t *testing.T) {
	dir := t.TempDir()
	crtFile := filepath.Join(dir, "crt.pem")
	keyFile := filepath.Join(dir, "key.pem")
	caFile := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(crtFile, []byte("crt-data\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, []byte("key-data\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caFile, []byte("ca-data\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cert := &Cert{
		CrtFile: crtFile,
		KeyFile: keyFile,
		CaFile:  caFile,
	}
	if err := cert.LoadData(); err != nil {
		t.Fatal(err)
	}
	if cert.CrtData != "crt-data\n" || cert.KeyData != "key-data\n" || cert.CaData != "ca-data\n" {
		t.Fatalf("unexpected loaded cert content: %+v", cert)
	}
}
