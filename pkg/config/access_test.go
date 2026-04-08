package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/stretchr/testify/assert"
)

func TestAccessFlags(t *testing.T) {
	ap := Access{}
	os.Args = []string{
		"app",
		"-conf", "/etc/openlan/fake.json",
		"-alias", "fake",
	}
	ap.Parse()
	fmt.Println(ap)
	assert.Equal(t, "fake", ap.Alias, "be the same.")
	assert.Equal(t, "/etc/openlan/fake.json", ap.Conf, "be the same.")
}

func TestAccess(t *testing.T) {
	ap := Access{
		Username: "user0@fake",
	}
	ap.Correct()
	assert.Equal(t, libol.INFO, ap.Log.Verbose, "be the same.")
	assert.Equal(t, "tcp", ap.Protocol, "be the same.")
	assert.Equal(t, "", ap.Crypt.Algo, "be the same.")
	assert.Equal(t, "fake", ap.Network, "be the same.")

	ap.Crypt.Secret = "fake-pass"
	ap.Correct()
	assert.Equal(t, "xor", ap.Crypt.Algo, "be the same.")
}

func TestParseForwardRule(t *testing.T) {
	rule := ParseForwardRule("8.8.8.8/32 to 192.168.11.2")
	assert.Equal(t, "8.8.8.8/32", rule.Prefix, "be the same.")
	assert.Equal(t, "192.168.11.2", rule.To, "be the same.")

	legacy := ParseForwardRule("8.8.4.4")
	assert.Equal(t, "8.8.4.4", legacy.Prefix, "be the same.")
	assert.Equal(t, "", legacy.To, "be the same.")
}

func TestAccessForwardRules(t *testing.T) {
	ap := Access{
		Forward: []string{
			"8.8.8.8/32 to 192.168.11.2",
			"8.8.4.4/32",
			"   ",
		},
	}
	rules := ap.ForwardRules()
	assert.Len(t, rules, 2, "be the same.")
	assert.Equal(t, "8.8.8.8/32", rules[0].Prefix, "be the same.")
	assert.Equal(t, "192.168.11.2", rules[0].To, "be the same.")
	assert.Equal(t, "8.8.4.4/32", rules[1].Prefix, "be the same.")
	assert.Equal(t, "", rules[1].To, "be the same.")
}
