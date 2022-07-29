package network

import (
	"github.com/luscis/openlan/pkg/libol"
	"sync"
)

type VirtualTap struct {
	lock   sync.RWMutex
	kernC  int
	kernQ  chan []byte
	virtC  int
	virtQ  chan []byte
	master Bridger
	tenant string
	flags  uint
	cfg    TapConfig
	name   string
	ifMtu  int
	sts    DeviceStats
}

func NewVirtualTap(tenant string, c TapConfig) (*VirtualTap, error) {
	name := c.Name
	if name == "" {
		name = Taps.GenName()
	}
	tap := &VirtualTap{
		cfg:    c,
		tenant: tenant,
		name:   name,
		ifMtu:  1514,
	}
	Taps.Add(tap)
	return tap, nil
}

func (t *VirtualTap) Type() string {
	return ProviderVir
}

func (t *VirtualTap) Tenant() string {
	return t.tenant
}

func (t *VirtualTap) IsTun() bool {
	return t.cfg.Type == TUN
}

func (t *VirtualTap) Name() string {
	return t.name
}

func (t *VirtualTap) hasFlags(v uint) bool {
	return t.flags&v == v
}

func (t *VirtualTap) setFlags(v uint) {
	t.flags |= v
}

func (t *VirtualTap) clearFlags(v uint) {
	t.flags &= ^v
}

func (t *VirtualTap) Has(v uint) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.hasFlags(v)
}

func (t *VirtualTap) Write(p []byte) (int, error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("VirtualTap.Write: %s % x", t, p[:20])
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.hasFlags(UsUp) {
		return 0, libol.NewErr("notUp")
	}
	if t.virtC >= t.cfg.VirBuf {
		libol.Warn("VirtualTap.Write: buffer fully")
		t.sts.Drop++
		return 0, nil
	}
	t.virtC++
	t.virtQ <- p
	return len(p), nil
}

func (t *VirtualTap) Read(p []byte) (int, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	data := <-t.kernQ
	t.lock.Lock()
	t.kernC--
	t.lock.Unlock()
	return copy(p, data), nil
}

func (t *VirtualTap) Recv(p []byte) (int, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	data := <-t.virtQ
	t.lock.Lock()
	t.virtC--
	t.lock.Unlock()
	return copy(p, data), nil
}

func (t *VirtualTap) Send(p []byte) (int, error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("VirtualTap.Send: %s % x", t, p[:20])
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.hasFlags(UsUp) {
		return 0, libol.NewErr("notUp")
	}
	if t.kernC >= t.cfg.KernBuf {
		t.sts.Drop++
		libol.Warn("VirtualTap.Send: buffer fully")
		return 0, nil
	}
	t.kernC++
	t.kernQ <- p
	return len(p), nil
}

func (t *VirtualTap) Close() error {
	t.Down()
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.hasFlags(UsClose) {
		return nil
	}
	t.setFlags(UsClose)
	t.clearFlags(UsUp)
	if t.master != nil {
		_ = t.master.DelSlave(t.name)
		t.master = nil
	}
	Taps.Del(t.name)
	return nil
}

func (t *VirtualTap) Master() Bridger {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.master
}

func (t *VirtualTap) SetMaster(dev Bridger) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.master == nil {
		t.master = dev
	}
	return libol.NewErr("already to %s", t.master)
}

func (t *VirtualTap) Up() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.hasFlags(UsUp) {
		t.kernC = 0
		t.kernQ = make(chan []byte, t.cfg.KernBuf)
		t.virtC = 0
		t.virtQ = make(chan []byte, t.cfg.VirBuf)
		t.setFlags(UsUp)
	}
}

func (t *VirtualTap) Down() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.hasFlags(UsUp) {
		t.clearFlags(UsUp)
		close(t.kernQ)
		t.kernQ = nil
		close(t.virtQ)
		t.virtQ = nil
	}
}

func (t *VirtualTap) String() string {
	return t.name
}

func (t *VirtualTap) Mtu() int {
	return t.ifMtu
}
