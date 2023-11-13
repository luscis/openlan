package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEspSpecifies(t *testing.T) {
	spec := EspSpecifies{
		State: EspState{
			Local: "3.3.3.1",
			Crypt: "fake-crypt",
			Auth:  "fake-auth",
		},
		Members: []*EspMember{
			{
				Peer:    "1.1.1.0",
				Address: "1.1.1.1",
				Spi:     0x01,
				State: EspState{
					Remote: "3.3.3.3",
				},
			},
		},
	}
	spec.Correct()
	assert.Equal(t, spec.State.Local, spec.Members[0].State.Local, "be the same.")
	assert.Equal(t, spec.State.Crypt, spec.Members[0].State.Crypt, "be the same.")
	assert.Equal(t, spec.State.Auth, spec.Members[0].State.Auth, "be the same.")
}

func TestEspSpecifies_GetMember(t *testing.T) {
	spec := EspSpecifies{
		State: EspState{
			Local: "3.3.3.1",
			Crypt: "fake-crypt",
			Auth:  "fake-auth",
		},
		Members: []*EspMember{
			{
				Peer:    "1.1.1.0",
				Address: "1.1.1.1",
				Spi:     123,
				State: EspState{
					Remote: "3.3.3.3",
				},
			},
		},
	}
	spec.Correct()
	obj := spec.GetMember("spi:123")
	assert.Equal(t, spec.Members[0], obj, "be the same.")
	// Add
	{
		mem1 := &EspMember{
			Peer:    "1.1.1.0",
			Address: "1.1.1.2",
			Spi:     124,
			State: EspState{
				Remote: "3.3.3.4",
			},
		}
		spec.AddMember(mem1)
		spec.Correct()
		obj1 := spec.GetMember("spi:124")
		assert.Equal(t, mem1, obj1, "be the same.")
	}
	// Delete
	{
		spec.DelMember("spi:123")
		obj0 := spec.GetMember("spi:123")
		assert.Equal(t, (*EspMember)(nil), obj0, "be the same.")

		spec.DelMember("spi:124")
		obj1 := spec.GetMember("spi:124")
		assert.Equal(t, (*EspMember)(nil), obj1, "be the same.")
	}
}

func TestEspSpecifies_AddPolicy(t *testing.T) {
	mem := &EspMember{
		Peer:    "1.1.1.0",
		Address: "1.1.1.2",
		Spi:     124,
		State: EspState{
			Local:  "3.3.3.1",
			Remote: "3.3.3.3",
		},
	}
	mem.Correct()
	assert.Equal(t, 1, len(mem.Policies), "be the same.")
	{
		po := &EspPolicy{
			Dest: "192.1.0.0/24",
		}
		mem.AddPolicy(po)
		assert.Equal(t, 2, len(mem.Policies), "be the same.")
		mem.RemovePolicy(po.Dest)
		assert.Equal(t, 1, len(mem.Policies), "be the same.")
	}
}
