package network

import (
	"sync"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/water"
)

type KernelTap struct {
	lock   sync.Mutex
	device *water.Interface
	master Bridger
	tenant string
	name   string
	config TapConfig
	ipMtu  int
}

func NewKernelTap(tenant string, c TapConfig) (*KernelTap, error) {
	if c.Mtu == 0 {
		c.Mtu = 1500
	}
	if c.Name == "auto" {
		c.Name = Taps.GenName()
	}
	device, err := WaterNew(c)
	if err != nil {
		return nil, err
	}
	tap := &KernelTap{
		tenant: tenant,
		device: device,
		name:   device.Name(),
		config: c,
		ipMtu:  c.Mtu,
	}
	Taps.Add(tap)
	return tap, nil
}

func (t *KernelTap) Has(v uint) bool {
	switch v {
	case UsClose:
		return t.device == nil
	case UsUp:
		return t.device != nil
	}
	return false
}

func (t *KernelTap) Type() string {
	return ProviderKer
}

func (t *KernelTap) Tenant() string {
	return t.tenant
}

func (t *KernelTap) IsTun() bool {
	return t.config.Type == TUN
}

func (t *KernelTap) Name() string {
	return t.name
}

func (t *KernelTap) Read(p []byte) (int, error) {
	t.lock.Lock()
	if t.device == nil {
		t.lock.Unlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.Unlock()
	if n, err := t.device.Read(p); err == nil {
		return n, nil
	} else {
		return 0, err
	}
}

func (t *KernelTap) Write(p []byte) (int, error) {
	t.lock.Lock()
	if t.device == nil {
		t.lock.Unlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.Unlock()
	return t.device.Write(p)
}

func (t *KernelTap) Recv(p []byte) (int, error) {
	return t.Read(p)
}

func (t *KernelTap) Send(p []byte) (int, error) {
	return t.Write(p)
}

func (t *KernelTap) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	libol.Debug("KernelTap.Close %s", t.name)
	if t.device == nil {
		return nil
	}
	if t.master != nil {
		_ = t.master.DelSlave(t.name)
		t.master = nil
	}
	err := t.device.Close()
	Taps.Del(t.name)
	t.device = nil
	return err
}

func (t *KernelTap) Master() Bridger {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.master
}

func (t *KernelTap) SetMaster(dev Bridger) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.master == nil {
		t.master = dev
	}
	return libol.NewErr("already to %s", t.master)
}

func (t *KernelTap) Up() {
	t.lock.Lock()
	defer t.lock.Unlock()
	libol.Debug("KernelTap.Up %s", t.name)
	if _, err := LinkUp(t.name); err != nil {
		libol.Warn("KernelTap.Up %s: %s", t.name, err)
	}
}

func (t *KernelTap) Down() {
	t.lock.Lock()
	defer t.lock.Unlock()
	libol.Debug("KernelTap.Down %s", t.name)
	if _, err := LinkDown(t.name); err != nil {
		libol.Warn("KernelTap.Down %s: %s", t.name, err)
	}
}

func (t *KernelTap) String() string {
	return t.name
}

func (t *KernelTap) Mtu() int {
	return t.ipMtu
}

func (t *KernelTap) Stats() DeviceStats {
	return GetDevStats(t.name)
}
