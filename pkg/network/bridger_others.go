// +build !linux

package network

func NewBridger(provider, name string, ifMtu int) Bridger {
	// others platform not support linux bridge.
	return NewVirtualBridge(name, ifMtu)
}
