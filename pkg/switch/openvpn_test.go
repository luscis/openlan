package cswitch

import (
	"testing"

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
}
