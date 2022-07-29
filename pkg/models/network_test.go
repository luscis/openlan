package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNetworkEqual(t *testing.T) {
	assert.Equal(t, true, NetworkEqual(nil, nil), "be the same.")
	o := &Network{}
	assert.Equal(t, false, NetworkEqual(o, nil), "be the same.")
	n := &Network{}
	assert.Equal(t, false, NetworkEqual(nil, n), "be the same.")
	assert.Equal(t, true, NetworkEqual(n, n), "be the same.")
	o = &Network{
		IfAddr:  "192.168.1.1",
		Netmask: "255.255.0.0",
		Routes: []*Route{
			{Prefix: "0.0.0.0/24", NextHop: "1.1.1.1."},
		},
	}
	n = &Network{
		IfAddr:  "192.168.1.1",
		Netmask: "255.255.0.0",
		Routes: []*Route{
			{Prefix: "0.0.0.0/24", NextHop: "1.1.1.1."},
		},
	}
	assert.Equal(t, true, NetworkEqual(o, n), "be the same.")
	o = &Network{
		IfAddr:  "192.168.1.1",
		Netmask: "255.255.0.0",
		Routes:  []*Route{},
	}
	assert.Equal(t, false, NetworkEqual(o, n), "be the same.")
	assert.Equal(t, false, NetworkEqual(n, o), "be the same.")
	o = &Network{
		IfAddr:  "192.168.1.1",
		Netmask: "255.255.0.0",
		Routes: []*Route{
			{Prefix: "0.0.0.0/24", NextHop: "1.1.1.1."},
			{Prefix: "0.0.0.1/24", NextHop: "1.1.1.1."},
		},
	}
	assert.Equal(t, false, NetworkEqual(o, n), "be the same.")
	assert.Equal(t, false, NetworkEqual(n, o), "be the same.")
	o = &Network{
		IfAddr:  "192.168.1.1",
		Netmask: "255.255.0.0",
		Routes: []*Route{
			{Prefix: "0.0.0.0/24", NextHop: "1.1.1.1."},
		},
	}
	assert.Equal(t, true, NetworkEqual(o, n), "be the same.")
	assert.Equal(t, true, NetworkEqual(n, o), "be the same.")
	o.IfAddr = "182.168.1.1"
	assert.Equal(t, false, NetworkEqual(o, n), "be the same.")
	assert.Equal(t, false, NetworkEqual(n, o), "be the same.")
	o.IfAddr = "192.168.1.1"
	assert.Equal(t, true, NetworkEqual(o, n), "be the same.")
	o.IfAddr = "255.255.255.0"
	assert.Equal(t, false, NetworkEqual(n, o), "be the same.")
}
