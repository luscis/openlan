package network

import (
	"errors"
	"github.com/luscis/openlan/pkg/libol"
	"net"
	"sync"
	"time"
)

type VirtualBridge struct {
	ipMtu   int
	name    string
	lock    sync.RWMutex
	ports   map[string]Taper
	macs    map[string]*MacFdb
	done    chan bool
	ticker  *time.Ticker
	timeout int
	address string
	kernel  Taper
	out     *libol.SubLogger
	sts     DeviceStats
}

func NewVirtualBridge(name string, mtu int) *VirtualBridge {
	b := &VirtualBridge{
		name:    name,
		ipMtu:   mtu,
		ports:   make(map[string]Taper, 1024),
		macs:    make(map[string]*MacFdb, 1024),
		done:    make(chan bool),
		ticker:  time.NewTicker(5 * time.Second),
		timeout: 5 * 60,
		out:     libol.NewSubLogger(name),
	}
	Bridges.Add(b)
	return b
}

func (b *VirtualBridge) Open(addr string) {
	tapCfg := TapConfig{
		Type: TAP,
		Mtu:  b.ipMtu,
	}
	b.out.Info("VirtualBridge.Open %s", addr)
	libol.Go(b.Start)
	if tap, err := NewKernelTap("", tapCfg); err != nil {
		b.out.Error("VirtualBridge.Open new kernel %s", err)
	} else {
		out, err := libol.IpLinkUp(tap.Name())
		if err != nil {
			b.out.Error("VirtualBridge.Open IpAddr %s:%s", err, out)
		}
		b.kernel = tap
		b.out.Info("VirtualBridge.Open %s", tap.Name())
		_ = b.AddSlave(tap.name)
	}
	if addr != "" && b.kernel != nil {
		b.address = addr
		if out, err := libol.IpAddrAdd(b.kernel.Name(), b.address); err != nil {
			b.out.Error("VirtualBridge.Open IpAddr %s:%s", err, out)
		}
	}
}

func (b *VirtualBridge) Kernel() string {
	if b.kernel == nil {
		return ""
	}
	return b.kernel.Name()
}

func (b *VirtualBridge) Close() error {
	if b.kernel != nil {
		if b.address != "" {
			out, err := libol.IpAddrDel(b.kernel.Name(), b.address)
			if err != nil {
				b.out.Error("VirtualBridge.Close: IpAddr %s:%s", err, out)
			}
		}
		_ = b.kernel.Close()
	}
	b.ticker.Stop()
	b.done <- true
	return nil
}

func (b *VirtualBridge) AddSlave(name string) error {
	tap := Taps.Get(name)
	if tap == nil {
		return libol.NewErr("%s notFound", name)
	}
	_ = tap.SetMaster(b)
	b.lock.Lock()
	b.ports[name] = tap
	b.lock.Unlock()
	b.out.Info("VirtualBridge.AddSlave: %s", name)
	libol.Go(func() {
		for {
			data := make([]byte, b.ipMtu)
			n, err := tap.Recv(data)
			if err != nil || n == 0 {
				break
			}
			if libol.HasLog(libol.DEBUG) {
				libol.Debug("VirtualBridge.KernelTap: %s % x", tap.Name(), data[:20])
			}
			m := &Framer{Data: data[:n], Source: tap}
			_ = b.Input(m)
		}
	})
	return nil
}

func (b *VirtualBridge) DelSlave(name string) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.ports[name]; ok {
		delete(b.ports, name)
	}
	b.out.Info("VirtualBridge.DelSlave: %s", name)
	return nil
}

func (b *VirtualBridge) ListSlave() <-chan Taper {
	data := make(chan Taper, 32)
	go func() {
		b.lock.RLock()
		defer b.lock.RUnlock()
		for _, obj := range b.ports {
			data <- obj
		}
		data <- nil
	}()
	return data
}

func (b *VirtualBridge) Type() string {
	return ProviderVir
}

func (b *VirtualBridge) String() string {
	return b.name
}

func (b *VirtualBridge) Name() string {
	return b.name
}

func (b *VirtualBridge) Forward(m *Framer) error {
	if err := b.UniCast(m); err != nil {
		_ = b.Flood(m)
	}
	return nil
}

func (b *VirtualBridge) Expire() error {
	deletes := make([]string, 0, 1024)
	//collect need deleted.
	b.lock.RLock()
	for index, learn := range b.macs {
		now := time.Now().Unix()
		if now-learn.Uptime > int64(b.timeout) {
			deletes = append(deletes, index)
		}
	}
	b.lock.RUnlock()
	b.out.Debug("VirtualBridge.Expire delete %d", len(deletes))
	//execute delete.
	b.lock.Lock()
	for _, d := range deletes {
		if _, ok := b.macs[d]; ok {
			delete(b.macs, d)
			b.out.Event("VirtualBridge.Expire: delete %s", d)
		}
	}
	b.lock.Unlock()
	return nil
}

func (b *VirtualBridge) Start() {
	libol.Go(func() {
		for {
			select {
			case <-b.done:
				return
			case t := <-b.ticker.C:
				b.out.Log("VirtualBridge.Start: Tick at %s", t)
				_ = b.Expire()
			}
		}
	})
}

func (b *VirtualBridge) Input(m *Framer) error {
	b.sts.Recv++
	b.Learn(m)
	return b.Forward(m)
}

func (b *VirtualBridge) Eth2Str(addr []byte) string {
	if len(addr) < 6 {
		return ""
	}
	return net.HardwareAddr(addr).String()
}

func (b *VirtualBridge) Learn(m *Framer) {
	mac := m.Data[6:12]
	if mac[0]&0x01 == 0x01 {
		return
	}
	key := b.Eth2Str(mac)
	if l := b.GetMac(key); l != nil {
		b.UpdateMac(key, m.Source)
		return
	}
	learn := &MacFdb{
		Device:  m.Source,
		Uptime:  time.Now().Unix(),
		NewTime: time.Now().Unix(),
		Address: make([]byte, 6),
	}
	copy(learn.Address, mac)
	b.out.Event("VirtualBridge.Learn: %s on %s", key, m.Source)
	b.AddMac(key, learn)
}

func (b *VirtualBridge) GetMac(mac string) *MacFdb {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if l, ok := b.macs[mac]; ok {
		return l
	}
	return nil
}

func (b *VirtualBridge) AddMac(mac string, fdb *MacFdb) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.macs[mac] = fdb
}

func (b *VirtualBridge) UpdateMac(mac string, device Taper) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if fdb, ok := b.macs[mac]; ok {
		fdb.Uptime = time.Now().Unix()
		fdb.Device = device
	}
}

func (b *VirtualBridge) ListMac() <-chan *MacFdb {
	data := make(chan *MacFdb, 32)
	go func() {
		b.lock.RLock()
		defer b.lock.RUnlock()
		for _, obj := range b.macs {
			data <- obj
		}
		data <- nil
	}()
	return data
}

func (b *VirtualBridge) Flood(m *Framer) error {
	data := m.Data
	from := m.Source
	if b.out.Has(libol.FLOW) {
		b.out.Flow("VirtualBridge.Flood: % x", data[:20])
	}
	outs := make([]Taper, 0, 32)
	b.lock.RLock()
	for _, port := range b.ports {
		if from != port {
			outs = append(outs, port)
		}
	}
	b.lock.RUnlock()
	for _, port := range outs {
		if b.out.Has(libol.FLOW) {
			b.out.Flow("VirtualBridge.Flood: %s % x", port, data[:20])
		}
		b.sts.Send++
		if _, err := port.Send(data); err != nil {
			b.out.Error("VirtualBridge.Flood: %s %s", port, err)
		}
	}
	return nil
}

func (b *VirtualBridge) UniCast(m *Framer) error {
	data := m.Data
	from := m.Source
	dest := b.Eth2Str(data[:6])
	learn := b.GetMac(dest)
	if learn == nil {
		return errors.New(dest + " notFound")
	}
	out := learn.Device
	if out != from && out.Has(UsUp) { // out should running
		b.sts.Send++
		if _, err := out.Send(data); err != nil {
			b.out.Warn("VirtualBridge.UniCast: %s %s", out, err)
		}
	} else {
		b.sts.Drop++
	}
	if b.out.Has(libol.FLOW) {
		b.out.Flow("VirtualBridge.UniCast: %s to %s % x", from, out, data[:20])
	}
	return nil
}

func (b *VirtualBridge) Mtu() int {
	return b.ipMtu
}

func (b *VirtualBridge) Stp(enable bool) error {
	return libol.NewErr("operation notSupport")
}

func (b *VirtualBridge) Delay(value int) error {
	return libol.NewErr("operation notSupport")
}

func (b *VirtualBridge) Stats() DeviceStats {
	return b.sts
}

func (b *VirtualBridge) CallIptables(value int) error {
	return libol.NewErr("operation notSupport")
}
