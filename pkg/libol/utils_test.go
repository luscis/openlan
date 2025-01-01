package libol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyTime(t *testing.T) {
	var s string

	s = PrettyTime(59)
	assert.Equal(t, "0m59s", s, "be the same.")
	s = PrettyTime(60 + 59)
	assert.Equal(t, "1m59s", s, "be the same.")

	s = PrettyTime(60*2 + 8)
	assert.Equal(t, "2m8s", s, "be the same.")

	s = PrettyTime(3600 + 1)
	assert.Equal(t, "1h0m", s, "be the same.")

	s = PrettyTime(3600 + 61)
	assert.Equal(t, "1h1m", s, "be the same.")

	s = PrettyTime(3600 + 60*59)
	assert.Equal(t, "1h59m", s, "be the same.")

	s = PrettyTime(3600*23 + 60*59)
	assert.Equal(t, "23h59m", s, "be the same.")

	s = PrettyTime(86400)
	assert.Equal(t, "1d0h", s, "be the same.")

	s = PrettyTime(86400 + 3600*5 + 59)
	assert.Equal(t, "1d5h", s, "be the same.")

	s = PrettyTime(86400 + 3600*23 + 59)
	assert.Equal(t, "1d23h", s, "be the same.")
}

func TestPrettyBytes(t *testing.T) {
	var s string

	s = PrettyBytes(1023)
	assert.Equal(t, "1023B", s, "be the same.")
	s = PrettyBytes(1024 + 1023)
	assert.Equal(t, "1.99K", s, "be the same.")

	s = PrettyBytes(1024*2 + 8)
	assert.Equal(t, "2.00K", s, "be the same.")

	s = PrettyBytes(1024*2 + 1023)
	assert.Equal(t, "2.99K", s, "be the same.")

	s = PrettyBytes(1024*1024 + 1)
	assert.Equal(t, "1.00M", s, "be the same.")

	s = PrettyBytes(1024*1024 + 1024*256 + 1023)
	assert.Equal(t, "1.25M", s, "be the same.")

	s = PrettyBytes(1024*1024 + 1024*1023)
	assert.Equal(t, "1.99M", s, "be the same.")

	s = PrettyBytes(1024 * 1024 * 1024)
	assert.Equal(t, "1.00G", s, "be the same.")

	s = PrettyBytes(1024*1024*1024 + 1024*1024*5 + 59)
	assert.Equal(t, "1.00G", s, "be the same.")

	s = PrettyBytes(1024*1024*1024 + 1024*1024*512 + 59)
	assert.Equal(t, "1.50G", s, "be the same.")
}
