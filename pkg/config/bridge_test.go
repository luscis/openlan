package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBridge(t *testing.T) {
	br := &Bridge{
		Network: "123",
	}
	br.Correct()
	assert.Equal(t, "br-123", br.Name, "be the same.")

	br1 := &Bridge{
		Network: "1234567890123456.cc",
	}
	br1.Correct()
	assert.Equal(t, "br-123456789012", br1.Name, "be the same.")
}
