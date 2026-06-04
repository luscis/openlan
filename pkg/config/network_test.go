package config

import "testing"

func TestNetworkUpdateNextHop(t *testing.T) {
	network := &Network{
		Routes: []PrefixRoute{
			{
				Prefix:  "192.0.2.0/24",
				NextHop: "192.168.66.1",
			},
			{
				Prefix: "198.51.100.0/24",
				MultiPath: []MultiPath{
					{NextHop: "192.168.66.1", Weight: 1},
					{NextHop: "192.168.67.1", Weight: 1},
				},
			},
			{
				Prefix:  "203.0.113.0/24",
				NextHop: "192.168.68.1",
			},
		},
	}

	if updated := network.UpdateNextHop("192.168.66.1", "192.168.66.2"); updated != 2 {
		t.Fatalf("expected 2 nexthops to be updated, got %d", updated)
	}

	if network.Routes[0].NextHop != "192.168.66.2" {
		t.Fatalf("unexpected route nexthop: %s", network.Routes[0].NextHop)
	}
	if network.Routes[1].MultiPath[0].NextHop != "192.168.66.2" {
		t.Fatalf("unexpected multipath nexthop: %s", network.Routes[1].MultiPath[0].NextHop)
	}
	if network.Routes[1].MultiPath[1].NextHop != "192.168.67.1" {
		t.Fatalf("unexpected unrelated multipath nexthop: %s", network.Routes[1].MultiPath[1].NextHop)
	}
	if network.Routes[2].NextHop != "192.168.68.1" {
		t.Fatalf("unexpected unrelated route nexthop: %s", network.Routes[2].NextHop)
	}
}

func TestNetworkSetAddressReplacesRouteNextHop(t *testing.T) {
	network := &Network{
		Bridge: &Bridge{Address: "192.168.66.1/24"},
		Routes: []PrefixRoute{
			{
				Prefix:  "192.0.2.0/24",
				NextHop: "192.168.66.1",
			},
		},
	}

	if updated := network.SetAddress("192.168.66.2/24"); updated != 1 {
		t.Fatalf("expected 1 nexthop to be updated, got %d", updated)
	}

	if network.Bridge.Address != "192.168.66.2/24" {
		t.Fatalf("unexpected bridge address: %s", network.Bridge.Address)
	}
	if network.Routes[0].NextHop != "192.168.66.2" {
		t.Fatalf("unexpected route nexthop: %s", network.Routes[0].NextHop)
	}
}
