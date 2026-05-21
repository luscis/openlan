package cswitch

import (
	"testing"
	"time"

	co "github.com/luscis/openlan/pkg/config"
)

func TestPeerName(t *testing.T) {
	in, out := PeerName("eth0", "-")
	if in != "eth0-i" || out != "eth0-o" {
		t.Fatalf("unexpected peer names: in=%q out=%q", in, out)
	}
}

func TestNewOpenLANWorker(t *testing.T) {
	cfg := &co.Network{
		Alias: "demo",
		Name:  "net-a",
	}
	w := NewOpenLANWorker(cfg)
	if w == nil || w.WorkerImpl == nil {
		t.Fatalf("expected worker to be initialized")
	}
	if w.alias != "demo" {
		t.Fatalf("unexpected alias: %q", w.alias)
	}
	if w.cfg != cfg {
		t.Fatalf("expected worker config to match input config")
	}
	if w.links == nil {
		t.Fatalf("expected links to be initialized")
	}
	if w.newTime <= 0 {
		t.Fatalf("expected newTime to be initialized")
	}
	if w.startTime != 0 {
		t.Fatalf("expected startTime to be zero before start")
	}
}

func TestOpenLANWorkerUpTime(t *testing.T) {
	w := &OpenLANWorker{}
	if got := w.UpTime(); got != 0 {
		t.Fatalf("expected uptime 0 when not started, got %d", got)
	}

	w.startTime = time.Now().Unix() - 3
	if got := w.UpTime(); got < 1 {
		t.Fatalf("expected positive uptime when started, got %d", got)
	}
}

func TestGetSwitchTransports(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantStream   string
		wantPacket   string
	}{
		{name: "default", input: "", wantStream: "tcp", wantPacket: "udp"},
		{name: "tls and kcp", input: "tls,kcp", wantStream: "tls", wantPacket: "kcp"},
		{name: "ssl alias", input: "ssl,udp", wantStream: "tls", wantPacket: "udp"},
		{name: "wss with spaces", input: " wss , kcp ", wantStream: "wss", wantPacket: "kcp"},
		{name: "unknown keeps default", input: "foo,bar", wantStream: "tcp", wantPacket: "udp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, packet := getSwitchTransports(tt.input)
			if stream != tt.wantStream || packet != tt.wantPacket {
				t.Fatalf("unexpected transports for %q: stream=%q packet=%q", tt.input, stream, packet)
			}
		})
	}
}
