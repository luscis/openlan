package cache

import (
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Network_LeaseAdd(t *testing.T) {
	// 0
	Network.AddLease("fake-alias", "192.168.1.1", "fake-net")
	{
		leAddr := Network.GetLeaseByAddr("192.168.1.1", "fake-net")
		assert.Equal(t, "fake-alias", leAddr.Alias, "MUST be found")
		leAlias := Network.GetLease("fake-alias", "fake-net")
		assert.Equal(t, "192.168.1.1", leAlias.Address, "MUST be found")
	}
	Network.DelLease("fake-alias", "fake-net")
	{
		leAddr := Network.GetLeaseByAddr("192.168.1.1", "fake-net")
		assert.Equal(t, (*schema.Lease)(nil), leAddr, "MUST be not found")
		leAlias := Network.GetLease("fake=alias", "fake-net")
		assert.Equal(t, (*schema.Lease)(nil), leAlias, "MUST be not found")
	}
	Network.AddLease("fake-aa", "192.168.1.1", "fake-aa")
	Network.AddLease("fake-cc", "192.168.1.1", "fake-aa")
	Network.DelLease("fake-aa", "fake-net")
	Network.DelLease("fake-cc", "fake-net")
	{
		leAddr := Network.GetLeaseByAddr("192.168.1.1", "fake-net")
		assert.Equal(t, (*schema.Lease)(nil), leAddr, "MUST be not found")
		leAlias := Network.GetLease("fake=alias", "fake-net")
		assert.Equal(t, (*schema.Lease)(nil), leAlias, "MUST be not found")
	}
	// 1
	Network.AddLease("fake-aa", "192.168.1.1", "fake-aa")
	Network.AddLease("fake-aa", "192.168.1.1", "fake-cc")
	{
		lea := Network.GetLease("fake-aa", "fake-aa")
		assert.Equal(t, "192.168.1.1", lea.Address, "MUST be found")
		lec := Network.GetLease("fake-aa", "fake-cc")
		assert.Equal(t, "192.168.1.1", lec.Address, "MUST be found")
	}
	n0 := &models.Network{
		IpStart: "192.168.1.0",
		IpEnd:   "192.168.1.222",
		Name:    "fake-aa",
	}
	Network.Add(n0)
	Network.NewLease("fake-vv", n0.Name)
	Network.NewLease("fake-dd", n0.Name)
	{
		le1 := Network.GetLease("fake-vv", n0.Name)
		assert.Equal(t, "192.168.1.0", le1.Address, "MUST be .0")
		le2 := Network.GetLease("fake-dd", n0.Name)
		assert.Equal(t, "192.168.1.2", le2.Address, "MUST be .2")
	}
	// 2
	n1 := &models.Network{
		IpStart: "192.168.1.0",
		IpEnd:   "192.168.1.222",
		Name:    "fake-dd",
	}
	Network.Add(n1)
	Network.NewLease("fake-vv", n1.Name)
	Network.NewLease("fake-dd", n1.Name)
	{
		le1 := Network.GetLease("fake-vv", n1.Name)
		assert.Equal(t, "192.168.1.0", le1.Address, "MUST be .0")
		le2 := Network.GetLease("fake-dd", n1.Name)
		assert.Equal(t, "192.168.1.1", le2.Address, "MUST be .1")
	}
}
