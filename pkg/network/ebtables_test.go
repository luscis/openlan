package network

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEBRuleArgsTCP(t *testing.T) {
	rule := EBRule{
		Source:  "192.61.0.1",
		Dest:    "10.254.0.12",
		Proto:   "tcp",
		DstPort: "80",
		Jump:    "drop",
	}

	args := strings.Join(rule.Args(), " ")

	assert.Equal(t, "-p IPv4 --ip-src 192.61.0.1 --ip-dst 10.254.0.12 --ip-proto tcp --ip-dport 80 -j DROP", args)
}

func TestEBRuleArgsDefaultActionIPv4Only(t *testing.T) {
	rule := EBRule{
		Jump: "accept",
	}

	args := strings.Join(rule.Args(), " ")

	assert.Equal(t, "-p IPv4 -j ACCEPT", args)
}

func TestEBRuleArgsHook(t *testing.T) {
	rule := EBRule{
		LogicalIn: "br-example",
		Jump:      "AT_example",
	}

	args := strings.Join(rule.Args(), " ")

	assert.Equal(t, "--logical-in br-example -j AT_example", args)
}
