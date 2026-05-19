package cswitch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestOpenVPN_Version(t *testing.T) {
	v1 := `OpenVPN 2.4.12 x86_64-redhat-linux-gnu [Fedora EPEL patched] [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on Mar 17 2022
	library versions: OpenSSL 1.0.2k-fips  26 Jan 2017, LZO 2.06
	Originally developed by James Yonan
	Copyright (C) 2002-2018 OpenVPN Inc <sales@openvpn.net>`

	vi := parseOpenVPNVersion(v1)
	assert.Equal(t, 24, vi, "notEqual")

	v2 := `OpenVPN 2.5.1 x86_64-pc-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO] [AEAD] built on May 14 2021
	library versions: OpenSSL 1.1.1w  11 Sep 2023, LZO 2.10
	Originally developed by James Yonan
	Copyright (C) 2002-2018 OpenVPN Inc <sales@openvpn.net>`
	vi = parseOpenVPNVersion(v2)
	assert.Equal(t, 25, vi, "notEqual")

	v3 := `something else`
	vi = parseOpenVPNVersion(v3)
	assert.Equal(t, 0, vi, "notEqual")
}

func TestNewOpenVPN_ListenParsing(t *testing.T) {
	cfg := &co.OpenVPN{
		Network:  "demo",
		Listen:   "10.20.30.40:1195",
		Protocol: "tcp",
		Device:   "tun1195",
		Subnet:   "10.119.0.0/24",
	}
	obj := NewOpenVPN(cfg)
	assert.Equal(t, "10.20.30.40", obj.Local)
	assert.Equal(t, "1195", obj.Port)
	assert.Equal(t, "tcp1195", obj.ID())

	cfg.Listen = ":4494"
	obj = NewOpenVPN(cfg)
	assert.Equal(t, "0.0.0.0", obj.Local)
	assert.Equal(t, "4494", obj.Port)

	cfg.Listen = "0.0.0.0"
	obj = NewOpenVPN(cfg)
	assert.Equal(t, "0.0.0.0", obj.Local)
	assert.Equal(t, "4494", obj.Port)
}

func TestOpenVPN_FileMethods(t *testing.T) {
	dir := t.TempDir()
	cfg := &co.OpenVPN{
		Network:   "demo",
		Listen:    "0.0.0.0:1194",
		Protocol:  "udp",
		Device:    "tun1194",
		Subnet:    "10.119.0.0/24",
		Directory: dir,
	}
	obj := NewOpenVPN(cfg)
	assert.Equal(t, "udp1194server.conf", obj.FileCfg(false))
	assert.Equal(t, filepath.Join(dir, "udp1194server.conf"), obj.FileCfg(true))
	assert.Equal(t, filepath.Join(dir, "udp1194server.pid"), obj.FilePid(true))
	assert.Equal(t, filepath.Join(dir, "ccd"), obj.ClientDir())
	assert.Equal(t, dir, obj.ServerDir())
}

func TestNewOpenVPNDataFromConf(t *testing.T) {
	cfg := &co.OpenVPN{
		Network:   "demo",
		Listen:    "0.0.0.0:1194",
		Protocol:  "udp",
		Device:    "tun1194",
		Subnet:    "10.119.0.0/24",
		Directory: t.TempDir(),
		RootCa:    "/tmp/ca.crt",
		ServerCrt: "/tmp/server.crt",
		ServerKey: "/tmp/server.key",
		DhPem:     "/tmp/dh.pem",
		TlsAuth:   "/tmp/ta.key",
		Script:    "/usr/bin/openlan -l test user check",
		Renego:    3600,
		Routes:    []string{"192.168.1.0/24", "not-a-cidr", "10.0.0.0/16"},
		Push:      []string{"redirect-gateway def1"},
		Cipher:    "AES-128-GCM:AES-256-GCM",
		Version:   25,
	}
	obj := NewOpenVPN(cfg)
	data := NewOpenVPNDataFromConf(obj)

	assert.Equal(t, "0.0.0.0", data.Local)
	assert.Equal(t, "1194", data.Port)
	assert.Equal(t, "udp", data.Protocol)
	assert.Equal(t, "tun1194", data.Device)
	assert.Equal(t, false, data.CertNot)
	assert.Equal(t, "10.119.0.0 255.255.255.0", data.Server)
	assert.Equal(t, []string{"192.168.1.0 24", "10.0.0.0 16"}, data.Routes)
	assert.Equal(t, []string{"redirect-gateway def1"}, data.Push)
	assert.Equal(t, "AES-128-GCM:AES-256-GCM", data.Cipher)
	assert.Equal(t, filepath.Join(cfg.Directory, "ccd"), data.ClientDir)
	assert.Equal(t, cfg.Directory, data.ServerDir)
}

func TestNewOpenVPNProfileFromConf(t *testing.T) {
	dir := t.TempDir()
	ca := filepath.Join(dir, "ca.crt")
	ta := filepath.Join(dir, "ta.key")
	assert.NoError(t, os.WriteFile(ca, []byte("CA-CONTENT\n"), 0600))
	assert.NoError(t, os.WriteFile(ta, []byte("TA-CONTENT\n"), 0600))

	cfg := &co.OpenVPN{
		Network:  "demo",
		Listen:   "127.0.0.1:1194",
		Protocol: "tcp",
		Device:   "tun1194",
		Subnet:   "10.119.0.0/24",
		RootCa:   ca,
		TlsAuth:  ta,
		Renego:   1800,
		Cipher:   "AES-128-GCM",
	}
	obj := NewOpenVPN(cfg)
	profile := NewOpenVPNProfileFromConf(obj)

	assert.Equal(t, "127.0.0.1", profile.Server)
	assert.Equal(t, "1194", profile.Port)
	assert.Equal(t, "tun", profile.Device)
	assert.Equal(t, "tcp", profile.Protocol)
	assert.Equal(t, 1800, profile.Renego)
	assert.Equal(t, "AES-128-GCM", profile.Cipher)
	assert.Equal(t, "CA-CONTENT\n", profile.Ca)
	assert.Equal(t, "TA-CONTENT\n", profile.TlsAuth)
}

func TestOpenVPNProfileRenderContainsCipher(t *testing.T) {
	dir := t.TempDir()
	ca := filepath.Join(dir, "ca.crt")
	ta := filepath.Join(dir, "ta.key")
	assert.NoError(t, os.WriteFile(ca, []byte("CA-CONTENT\n"), 0600))
	assert.NoError(t, os.WriteFile(ta, []byte("TA-CONTENT\n"), 0600))

	cfg := &co.OpenVPN{
		Network:  "demo",
		Listen:   "127.0.0.1:1194",
		Protocol: "tcp",
		Device:   "tun1194",
		Subnet:   "10.119.0.0/24",
		RootCa:   ca,
		TlsAuth:  ta,
		Cipher:   "AES-128-GCM:AES-256-GCM",
	}
	obj := NewOpenVPN(cfg)
	ctx, err := obj.Profile()
	assert.NoError(t, err)

	text := string(ctx)
	assert.Contains(t, text, "data-ciphers AES-128-GCM:AES-256-GCM")
	assert.True(t, strings.Contains(text, "<ca>") && strings.Contains(text, "</ca>"))
	assert.True(t, strings.Contains(text, "<tls-auth>") && strings.Contains(text, "</tls-auth>"))
}
