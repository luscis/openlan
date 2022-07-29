package network

func NewBridger(provider, name string, ifMtu int) Bridger {
	if provider == ProviderVir {
		return NewVirtualBridge(name, ifMtu)
	}
	return NewLinuxBridge(name, ifMtu)
}
