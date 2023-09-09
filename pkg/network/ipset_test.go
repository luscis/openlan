package network

import (
	"testing"
)

func TestIPSetCreateDestroy(t *testing.T) {
	i := NewIPSet("hi", "hash:net")
	if out, err := i.Create(); err != nil {
		t.Skipf(out)
	}
	if out, err := i.Add("192.168.0.0/24"); err != nil {
		t.Skipf(out)
	}
	if out, err := i.Add("192.168.1.0/24"); err != nil {
		t.Skipf(out)
	}
	if out, err := i.Add("192.168.2.0/24"); err != nil {
		t.Skipf(out)
	}
	if out, err := i.Del("192.168.1.0/24"); err != nil {
		t.Skipf(out)
	}
	if out, err := i.Flush(); err != nil {
		t.Skipf(out)
	}
	if out, err := i.Destroy(); err != nil {
		t.Skipf(out)
	}
}
