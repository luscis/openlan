package cswitch

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/luscis/openlan/pkg/schema"
)

func TestCeciWorkerIsRunning(t *testing.T) {
	dir := t.TempDir()
	pidfile := filepath.Join(dir, "ceci.pid")
	if err := os.WriteFile(pidfile, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		t.Fatal(err)
	}

	w := &CeciWorker{}
	if !w.isRunning(pidfile) {
		t.Fatalf("expected current process pid to be treated as running")
	}

	if err := os.WriteFile(pidfile, []byte("999999"), 0600); err != nil {
		t.Fatal(err)
	}
	if w.isRunning(pidfile) {
		t.Fatalf("expected stale pid to be treated as not running")
	}
}

func TestNormalizeCeciCertLoadsFileContent(t *testing.T) {
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

	cert, err := normalizeCeciCert(&schema.Cert{
		CrtFile:  crtFile,
		KeyFile:  keyFile,
		CaFile:   caFile,
		Insecure: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cert == nil {
		t.Fatalf("expected cert content to be loaded")
	}
	if cert.CrtData != "crt-data" || cert.KeyData != "key-data" || cert.CaData != "ca-data" {
		t.Fatalf("unexpected cert content: %+v", cert)
	}
	if !cert.Insecure {
		t.Fatalf("expected insecure flag to be preserved")
	}
}
