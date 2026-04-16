package proxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	co "github.com/luscis/openlan/pkg/config"
)

func TestHttpProxyIsAuthNetworkFilter(t *testing.T) {
	h := &HttpProxy{
		cfg: &co.HttpProxy{
			Network: "guest",
		},
		pass: map[string]string{
			"alice": "secret",
		},
	}

	if !h.isAuth("alice", "secret") {
		t.Fatalf("expected alice to authenticate")
	}
	if h.isAuth("alice@guest", "secret") {
		t.Fatalf("expected alice@guest to be rejected without exact username match")
	}
	if h.isAuth("bob", "secret") {
		t.Fatalf("expected bob to be rejected without exact username match")
	}
}

func TestHttpProxyIsAuthNoDefaultFallback(t *testing.T) {
	h := &HttpProxy{
		cfg: &co.HttpProxy{},
		pass: map[string]string{
			"alice":         "secret",
			"alice@default": "secret",
		},
	}

	if !h.isAuth("alice", "secret") {
		t.Fatalf("expected exact username match to authenticate")
	}
	if h.isAuth("bob", "secret") {
		t.Fatalf("expected no default-network fallback when network is unset")
	}
}

func TestHttpProxyRefreshPass(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "password")
	if err := os.WriteFile(file, []byte("alice@guest:old:guest:2026-12-19T00\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := NewHttpProxy(&co.HttpProxy{
		Password: file,
	}, nil)

	if !h.isAuth("alice", "old") {
		t.Fatalf("expected initial password to authenticate")
	}

	if err := os.WriteFile(file, []byte("alice@guest:new:guest:2026-12-19T00\nbob@guest:secret:guest:2026-12-19T00\n"), 0600); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(file, future, future); err != nil {
		t.Fatal(err)
	}

	h.loadPass()

	if h.isAuth("alice", "old") {
		t.Fatalf("expected old password to stop working after reload")
	}
	if !h.isAuth("alice", "new") {
		t.Fatalf("expected new password to work after reload")
	}
	if !h.isAuth("bob", "secret") {
		t.Fatalf("expected newly added user to authenticate after reload")
	}
}

func TestHttpProxyLoadPassNetworkFilter(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "password")
	if err := os.WriteFile(file, []byte("alice@guest:secret:guest:2026-12-19T00\nbob@admin:secret:admin:2026-12-19T00\ncarol:secret:guest:2026-12-19T00\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := NewHttpProxy(&co.HttpProxy{
		Password: file,
		Network:  "guest",
	}, nil)

	h.passLock.RLock()
	defer h.passLock.RUnlock()

	if _, ok := h.pass["alice"]; !ok {
		t.Fatalf("expected guest user to be loaded as short name")
	}
	if _, ok := h.pass["bob"]; ok {
		t.Fatalf("expected admin user to be filtered out")
	}
	if _, ok := h.pass["carol"]; ok {
		t.Fatalf("expected networkless user to be filtered out when network is set")
	}
}

func TestHttpProxyLoadPassIgnoresTrailingFields(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "password")
	if err := os.WriteFile(file, []byte("admin@admin:jy2q8zxg26od:guest:2026-12-19T00\ndaniel@admin:mnaf80sq3lye:guest:2028-01-20T00\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := NewHttpProxy(&co.HttpProxy{
		Password: file,
		Network:  "admin",
	}, nil)

	h.passLock.RLock()
	defer h.passLock.RUnlock()

	if !h.isAuth("admin", "jy2q8zxg26od") {
		t.Fatalf("expected password field to be loaded without trailing metadata")
	}
	if !h.isAuth("daniel", "mnaf80sq3lye") {
		t.Fatalf("expected second user to authenticate with trimmed password")
	}
	if h.isAuth("admin", "jy2q8zxg26od:guest:2026-12-19T00") {
		t.Fatalf("expected trailing fields not to be part of the stored password")
	}
}

func TestHttpProxyLoadPassKeepsRawUserWithoutNetwork(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "password")
	if err := os.WriteFile(file, []byte("alice@example:secret\n"), 0600); err != nil {
		t.Fatal(err)
	}

	h := NewHttpProxy(&co.HttpProxy{
		Password: file,
	}, nil)

	h.passLock.RLock()
	defer h.passLock.RUnlock()

	if _, ok := h.pass["alice"]; !ok {
		t.Fatalf("expected short username to be kept when network is unset")
	}
	if _, ok := h.pass["alice@example"]; ok {
		t.Fatalf("expected network info not to be stored in passmap")
	}
}

func TestHttpProxySaveStatsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "stats.json")
	h := &HttpProxy{
		cfg: &co.HttpProxy{
			StatsFile: file,
		},
		statsFile: file,
		requests:  make(map[string]*HttpRecord),
		startat:   time.Now().Add(-time.Minute),
	}
	h.requests["example.com"] = &HttpRecord{
		Domain: "example.com",
		Bytes:  123,
	}
	h.saveStats()

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "\"Total\": 1") {
		t.Fatalf("expected total count to be saved, got %s", text)
	}
	if !strings.Contains(text, "\"Bytes\": 123") {
		t.Fatalf("expected byte count to be saved, got %s", text)
	}
}
