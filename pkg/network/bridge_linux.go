package network

import (
	"fmt"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
	nl "github.com/vishvananda/netlink"
)

func NewBridger(provider, name string, ifMtu int) Bridger {
	return NewLinuxBridge(name, ifMtu)
}

type LinuxBridge struct {
	sts     DeviceStats
	address *nl.Addr
	ipMtu   int
	name    string
	device  nl.Link
	ctl     *BrCtl
	out     *libol.SubLogger
	l3if    string
	l2if    string
}

func GetPair(name string) (string, string) {
	s0 := ""
	s1 := ""
	if strings.HasPrefix(name, "br-") {
		s0 = strings.Replace(name, "br-", "hi-", 1)
		s1 = strings.Replace(name, "br-", "si-", 1)
	} else {
		s0 = fmt.Sprintf("hi-%s", name)
		s1 = fmt.Sprintf("si-%s", name)
	}
	return GetName(s0), GetName(s1)
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
	b.l3if, b.l2if = GetPair(name)

	Bridges.Add(b)
	return b
}

func (b *LinuxBridge) Kernel() string {
	return b.name
}

func (b *LinuxBridge) Open(addr string) {
	b.out.Debug("LinuxBridge.Open")
	link, _ := nl.LinkByName(b.name)
	if link == nil {
		br := &nl.Bridge{
			LinkAttrs: nl.LinkAttrs{
				TxQLen: -1,
				Name:   b.name,
				MTU:    b.ipMtu,
			},
		}
		err := nl.LinkAdd(br)
		if err != nil {
			b.out.Error("LinuxBridge.Open: %s", err)
			return
		}
		link, err = nl.LinkByName(b.name)
		if link == nil {
			b.out.Error("LinuxBridge.Open: %s", err)
			return
		}
	}
	if err := nl.LinkSetUp(link); err != nil {
		libol.Error("LinuxBridge.Open: %s", err)
	}

	b.out.Info("LinuxBridge.Open success")

	if addr != "" {
		ipAddr, err := nl.ParseAddr(addr)
		if err != nil {
			b.out.Error("LinuxBridge.Open: ParseAddr %s", err)
		}
		b.address = ipAddr
		if err := b.Plugin(ipAddr); err != nil {
			libol.Error("LinuxBridge.Open: %s", err)
		}
	}
	b.device = link
}

func (b *LinuxBridge) Close() error {
	var err error
	if b.address != nil {
		b.Unplugin()
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

func (b *LinuxBridge) Plugin(addr *nl.Addr) error {
	if link, _ := nl.LinkByName(b.l2if); link != nil {
		return nil
	}

	link := &nl.Veth{
		LinkAttrs: nl.LinkAttrs{Name: b.l3if},
		PeerName:  b.l2if,
	}
	if err := nl.LinkAdd(link); err != nil {
		return err
	}
	if err := nl.LinkSetUp(link); err != nil {
		return err
	}

	if err := b.AddSlave(b.l2if); err != nil {
		return err
	}
	if err := nl.AddrAdd(link, addr); err != nil {
		return err
	}

	return nil
}

func (b *LinuxBridge) Unplugin() error {
	link, _ := nl.LinkByName(b.l2if)
	if link == nil {
		return nil
	}

	if err := nl.LinkDel(link); err != nil {
		return err
	}

	return nil
}

func (b *LinuxBridge) L3Name() string {
	return b.l3if
}

func (b *LinuxBridge) SetMtu(mtu int) error {
	link, _ := nl.LinkByName(b.l3if)
	if link == nil {
		return nil
	}
	if err := nl.LinkSetMTU(link, mtu); err != nil {
		return err
	}

	link, _ = nl.LinkByName(b.l2if)
	if link == nil {
		return nil
	}
	if err := nl.LinkSetMTU(link, mtu); err != nil {
		return err
	}

	return nil
}
