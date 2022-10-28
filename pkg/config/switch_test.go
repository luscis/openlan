package config

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSwitch(t *testing.T) {
	sw := Switch{}
	sw.Correct()
	assert.Equal(t, libol.INFO, sw.Log.Verbose, "be the same.")
	assert.Equal(t, "0.0.0.0:10002", sw.Listen, "be the same.")
	assert.Equal(t, "0.0.0.0:10000", sw.Http.Listen, "be the same.")
	sw.Listen = "192.168.1.0"
	sw.Correct()
	assert.Equal(t, "192.168.1.0:10002", sw.Listen, "be the same.")
}
