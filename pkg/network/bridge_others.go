//go:build !linux

package network

type OtherBridge struct {
	name string
	mtu  int
}

func NewBridger(provider, name string, ifMtu int) Bridger {
	return &OtherBridge{
		name: name,
		mtu:  ifMtu,
	}
}

func (b *OtherBridge) Type() string {
	return "NAN"
}

func (b *OtherBridge) Name() string {
	return b.name
}
func (b *OtherBridge) Open(addr string) {
}

func (b *OtherBridge) Close() error {
	return nil
}

func (b *OtherBridge) AddSlave(name string) error {
	return nil
}

func (b *OtherBridge) DelSlave(name string) error {
	return nil
}

func (b *OtherBridge) ListSlave() <-chan Taper {
	return nil
}

func (b *OtherBridge) Mtu() int {
	return b.mtu
}

func (b *OtherBridge) Stp(enable bool) error {
	return nil
}

func (b *OtherBridge) Delay(value int) error {
	return nil
}

func (b *OtherBridge) Kernel() string {
	return "NAN"
}

func (b *OtherBridge) ListMac() <-chan *MacFdb {
	return nil
}

func (b *OtherBridge) String() string {
	return "NAN"
}

func (b *OtherBridge) Stats() DeviceInfo {
	return DeviceInfo{}
}

func (b *OtherBridge) CallIptables(value int) error {
	return nil
}

func (b *OtherBridge) L3Name() string {
	return "NAN"
}

func (b *OtherBridge) SetMtu(mtu int) error {
	return nil
}
