package network

func NewBridger(provider, name string, ifMtu int) Bridger {
	return NewLinuxBridge(name, ifMtu)
}
