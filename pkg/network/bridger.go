package network

import (
	"sync"
)

const (
	ProviderVir = "virtual"
	ProviderKer = "kernel"
	ProviderLin = "linux"
)

type MacFdb struct {
	Address []byte
	Device  Taper
	Uptime  int64
	NewTime int64
}

type Bridger interface {
	Type() string
	Name() string
	Open(addr string)
	Close() error
	AddSlave(name string) error
	DelSlave(name string) error
	ListSlave() <-chan Taper
	Mtu() int
	Stp(enable bool) error
	Delay(value int) error
	Kernel() string // name in kernel.
	ListMac() <-chan *MacFdb
	String() string
	Stats() DeviceInfo
	CallIptables(value int) error
	L3Name() string
	SetMtu(mtu int) error
}

type bridger struct {
	lock    sync.RWMutex
	index   int
	devices map[string]Bridger
}

func (t *bridger) Add(br Bridger) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.devices == nil {
		t.devices = make(map[string]Bridger, 1024)
	}
	t.devices[br.Name()] = br
}

func (t *bridger) Get(name string) Bridger {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.devices == nil {
		return nil
	}
	if t, ok := t.devices[name]; ok {
		return t
	}
	return nil
}

func (t *bridger) Del(name string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.devices == nil {
		return
	}
	delete(t.devices, name)
}

func (t *bridger) List() <-chan Bridger {
	data := make(chan Bridger, 32)
	go func() {
		t.lock.RLock()
		defer t.lock.RUnlock()
		for _, obj := range t.devices {
			data <- obj
		}
		data <- nil
	}()
	return data
}

var Bridges = &bridger{}
