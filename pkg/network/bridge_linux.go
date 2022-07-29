package network

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/vishvananda/netlink"
)

type LinuxBridge struct {
	sts     DeviceStats
	address *netlink.Addr
	ipMtu   int
	name    string
	device  netlink.Link
	ctl     *BrCtl
	out     *libol.SubLogger
}

func NewLinuxBridge(name string, mtu int) *LinuxBridge {
	if mtu == 0 {
		mtu = 1500
	}
	b := &LinuxBridge{
		name:  name,
		ipMtu: mtu,
		ctl:   NewBrCtl(name, mtu),
		out:   libol.NewSubLogger(name),
	}
	Bridges.Add(b)
	return b
}

func (b *LinuxBridge) Kernel() string {
	return b.name
}

func (b *LinuxBridge) Open(addr string) {
	b.out.Debug("LinuxBridge.Open")
	link, _ := netlink.LinkByName(b.name)
	if link == nil {
		br := &netlink.Bridge{
			LinkAttrs: netlink.LinkAttrs{
				TxQLen: -1,
				Name:   b.name,
				MTU:    b.ipMtu,
			},
		}
		err := netlink.LinkAdd(br)
		if err != nil {
			b.out.Error("LinuxBridge.Open: %s", err)
			return
		}
		link, err = netlink.LinkByName(b.name)
		if link == nil {
			b.out.Error("LinuxBridge.Open: %s", err)
			return
		}
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("LinuxBridge.Open: %s", err)
	}
	b.out.Info("LinuxBridge.Open success")
	if addr != "" {
		ipAddr, err := netlink.ParseAddr(addr)
		if err != nil {
			b.out.Error("LinuxBridge.Open: ParseCIDR %s", err)
		}
		if err := netlink.AddrAdd(link, ipAddr); err != nil {
			b.out.Error("LinuxBridge.Open: SetLinkIp: %s", err)
		}
		b.address = ipAddr
	}
	b.device = link
}

func (b *LinuxBridge) Close() error {
	var err error
	if b.device != nil && b.address != nil {
		if err = netlink.AddrDel(b.device, b.address); err != nil {
			b.out.Error("LinuxBridge.Close: UnsetLinkIp %s", err)
		}
	}
	return err
}

func (b *LinuxBridge) AddSlave(name string) error {
	if err := b.ctl.AddPort(name); err != nil {
		b.out.Error("LinuxBridge.AddSlave: %s", name)
		return err
	}
	b.out.Info("LinuxBridge.AddSlave: %s", name)
	return nil
}

func (b *LinuxBridge) DelSlave(name string) error {
	if err := b.ctl.DelPort(name); err != nil {
		b.out.Error("LinuxBridge.DelSlave: %s", name)
		return err
	}
	b.out.Info("LinuxBridge.DelSlave: %s", name)
	return nil
}

func (b *LinuxBridge) ListSlave() <-chan Taper {
	data := make(chan Taper, 32)
	go func() {
		data <- nil
	}()
	b.out.Warn("LinuxBridge.ListSlave: notSupport")
	return data
}

func (b *LinuxBridge) Type() string {
	return ProviderLin
}

func (b *LinuxBridge) String() string {
	return b.name
}

func (b *LinuxBridge) Name() string {
	return b.name
}

func (b *LinuxBridge) Mtu() int {
	return b.ipMtu
}

func (b *LinuxBridge) Stp(enable bool) error {
	return b.ctl.Stp(enable)
}

func (b *LinuxBridge) Delay(value int) error {
	return b.ctl.Delay(value)
}

func (b *LinuxBridge) ListMac() <-chan *MacFdb {
	data := make(chan *MacFdb, 32)
	go func() {
		data <- nil
	}()
	b.out.Warn("LinuxBridge.ListMac: notSupport")
	return data
}

func (b *LinuxBridge) Stats() DeviceStats {
	return b.sts
}

func (b *LinuxBridge) CallIptables(value int) error {
	return b.ctl.CallIptables(value)
}
