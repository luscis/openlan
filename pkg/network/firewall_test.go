package network

import (
	"testing"
)

var firewall = NewFireWallGlobal(nil)

func TestFireWallStart(t *testing.T) {
	firewall.Initialize()
	firewall.Start()
}

func TestFireWallTableFilter(t *testing.T) {
	tab := NewFireWallTable("fake")
	tab.Filter.In.AddRule(IpRule{
		Output: "br-fake_1",
		Input:  "br-fake_1",
		Source: "192.168.0.1/24",
		Dest:   "192.168.1.0/24",
	})

	tab.Filter.Install()

	tab_1 := NewFireWallTable("fake_1")
	tab_1.Filter.In.AddRule(IpRule{
		Output: "br-fake_1",
		Input:  "br-fake_1",
		Source: "192.168.1.1",
		Dest:   "192.168.3.0",
	})
	tab_1.Filter.For.AddRule(IpRule{
		Output: "br-fake_1",
		Input:  "br-fake_1",
		Source: "192.168.0.1",
		Dest:   "192.168.3.0/24",
	})
	tab_1.Filter.Install()

	tab.Filter.Cancel()
	tab_1.Filter.Cancel()
}

func TestFireWallTableNAT(t *testing.T) {
	tab := NewFireWallTable("fake")
	tab.Nat.Install()

	tab.Nat.Cancel()
}

func TestFireWallTableMangle(t *testing.T) {
	tab := NewFireWallTable("fake")
	tab.Mangle.Install()

	tab.Mangle.Cancel()
}

func TestFireWallTableRaw(t *testing.T) {
	tab := NewFireWallTable("fake")
	tab.Raw.Install()

	tab.Raw.Cancel()
}

func TestFireWallCancel(t *testing.T) {
	firewall.Stop()
}
