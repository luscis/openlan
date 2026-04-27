package config

import "testing"

func TestNormalizeOpenVPNCipher(t *testing.T) {
	value, err := NormalizeOpenVPNCipher("AES-128-GCM:aes-256-cbc:AES-128-GCM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "AES-128-GCM:AES-256-CBC" {
		t.Fatalf("unexpected value: %s", value)
	}
}

func TestNormalizeOpenVPNCipherRejectUnsupported(t *testing.T) {
	if _, err := NormalizeOpenVPNCipher("CHACHA20-POLY1305"); err == nil {
		t.Fatalf("expected unsupported cipher error")
	}
}
