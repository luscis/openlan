package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/stretchr/testify/assert"
)

func TestPointFlags(t *testing.T) {
	ap := Point{}
	os.Args = []string{
		"app",
		"-conf", "/etc/openlan/fake.json",
		"-terminal", "off",
		"-alias", "fake",
	}
	ap.Parse()
	fmt.Println(ap)
	assert.Equal(t, "fake", ap.Alias, "be the same.")
	assert.Equal(t, "/etc/openlan/fake.json", ap.Conf, "be the same.")
	assert.Equal(t, "off", ap.Terminal, "be the same.")
}

func TestPoint(t *testing.T) {
	ap := Point{
		Username: "user0@fake",
	}
	ap.Correct()
	assert.Equal(t, libol.INFO, ap.Log.Verbose, "be the same.")
	assert.Equal(t, "tcp", ap.Protocol, "be the same.")
	assert.Equal(t, "", ap.Crypt.Algo, "be the same.")
	assert.Equal(t, "fake", ap.Network, "be the same.")
	assert.Equal(t, "on", ap.Terminal, "be the same.")

	ap.Crypt.Secret = "fake-pass"
	ap.Correct()
	assert.Equal(t, "xor", ap.Crypt.Algo, "be the same.")
}
