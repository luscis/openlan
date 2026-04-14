package cswitch

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
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
