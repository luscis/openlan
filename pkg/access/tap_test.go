package access

import (
	"testing"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/water"
)

func TestTapWrite(t *testing.T) {
	cfg := water.Config{DeviceType: water.TAP}
	dev, err := water.New(cfg)
	if err != nil {
		t.Skipf("Tap.open %s", err)
		return
	}
	if _, err := libol.IpLinkUp(dev.Name()); err != nil {
		libol.Warn("KernelTap.Up %s: %s", dev, err)
		return
	}
	frame := make([]byte, 65)
	for i := 0; i < 64; i++ {
		frame[i] = uint8(i)
	}
	n, err := dev.Write(frame)
	if err != nil {
		t.Errorf("Tap.write: %s", err)
	}
	if n != len(frame) {
		t.Errorf("Tap.write: %d", n)
	}
}

func BenchmarkTapWrite64(b *testing.B) {
	cfg := water.Config{DeviceType: water.TAP}
	dev, err := water.New(cfg)
	if err != nil {
		b.Skipf("Tap.open %s", err)
		return
	}
	if _, err := libol.IpLinkUp(dev.Name()); err != nil {
		libol.Warn("KernelTap.Up %s: %s", dev, err)
		return
	}
	//b.Logf("Tap.write: to %s", dev.Name())
	for i := 0; i < b.N; i++ {
		frame := make([]byte, 64)
		for i := 0; i < len(frame); i++ {
			frame[i] = uint8(i)
		}

		//b.Logf("Tap.write: frame %d", len(frame))
		n, err := dev.Write(frame)
		if err != nil {
			b.Errorf("Tap.write: %s", err)
		}
		if n != len(frame) {
			b.Errorf("Tap.write: %d", n)
		}
	}
}

func BenchmarkTapWrite1500(b *testing.B) {
	cfg := water.Config{DeviceType: water.TAP}
	dev, err := water.New(cfg)
	if err != nil {
		b.Skipf("Tap.open %s", err)
		return
	}
	if _, err := libol.IpLinkUp(dev.Name()); err != nil {
		libol.Warn("KernelTap.Up %s: %s", dev, err)
		return
	}
	//b.Logf("Tap.write: to %s", dev.Name())
	for i := 0; i < b.N; i++ {
		frame := make([]byte, 1500)
		for i := 0; i < len(frame); i++ {
			frame[i] = uint8(i)
		}

		//b.Logf("Tap.write: frame %d", len(frame))
		n, err := dev.Write(frame)
		if err != nil {
			b.Errorf("Tap.write: %s", err)
		}
		if n != len(frame) {
			b.Errorf("Tap.write: %d", n)
		}
	}
}
