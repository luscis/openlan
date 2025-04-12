package access

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
	"github.com/miekg/dns"
)

type TapWorkerListener struct {
	OnOpen   func(w *TapWorker) error
	OnClose  func(w *TapWorker)
	FindNext func(dest []byte) []byte
	ReadAt   func(frame *libol.FrameMessage) error
	OnDNS    func(string, net.IP)
}

type TunEther struct {
	HwAddr []byte
	IpAddr []byte
}

type TapWorker struct {
	// private
	lock       sync.Mutex
	device     network.Taper
	listener   TapWorkerListener
	ether      TunEther
	neighbor   Neighbors
	devCfg     network.TapConfig
	pinCfg     *config.Point
	ifAddr     string
	writeQueue chan *libol.FrameMessage
	done       chan bool
	out        *libol.SubLogger
	eventQueue chan *WorkerEvent
}

func NewTapWorker(devCfg network.TapConfig, pinCfg *config.Point) (a *TapWorker) {
	a = &TapWorker{
		devCfg:     devCfg,
		pinCfg:     pinCfg,
		done:       make(chan bool, 2),
		writeQueue: make(chan *libol.FrameMessage, pinCfg.Queue.TapWr),
		out:        libol.NewSubLogger(pinCfg.Id()),
		eventQueue: make(chan *WorkerEvent, 32),
	}
	return
}

func (a *TapWorker) Initialize() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.out.Info("TapWorker.Initialize")
	a.neighbor = Neighbors{
		neighbors: make(map[uint32]*Neighbor, 1024),
		done:      make(chan bool),
		ticker:    time.NewTicker(5 * time.Second),
		timeout:   3 * 60,
		interval:  60,
		listener: NeighborListener{
			Interval: func(dest []byte) {
				a.OnArpAlive(dest)
			},
			Expire: func(dest []byte) {
				a.OnArpAlive(dest)
			},
		},
	}
	if a.IsTun() {
		addr := a.pinCfg.Interface.Address
		a.setAddr(addr, libol.GenEthAddr(6))
		a.out.Info("TapWorker.Initialize: src %x", a.ether.HwAddr)
	}
	if err := a.open(); err != nil {
		a.eventQueue <- NewEvent(EvTapOpenErr, err.Error())
	}
}

func (a *TapWorker) IsTun() bool {
	return a.devCfg.Type == network.TUN
}

func (a *TapWorker) setIpAddr(ipaddr string) {
	// format ip address.
	if addr, err := libol.IPNetmask(ipaddr); err == nil {
		ifAddr := strings.SplitN(addr, "/", 2)[0]
		a.ether.IpAddr = net.ParseIP(ifAddr).To4()
		if a.ether.IpAddr == nil {
			a.ether.IpAddr = []byte{0x00, 0x00, 0x00, 0x00}
		}
		a.out.Info("TapWorker.setEther: srcIp % x", a.ether.IpAddr)
		// changed address need open device again.
		if a.ifAddr != "" && a.ifAddr != addr {
			a.out.Warn("TapWorker.setEther changed %s->%s", a.ifAddr, addr)
			a.eventQueue <- NewEvent(EvTapReset, "ifAddr changed")
		}
		a.ifAddr = addr
	} else {
		a.out.Warn("TapWorker.setEther: %s: %s", addr, err)
	}
}
func (a *TapWorker) setAddr(ipAddr string, hwAddr []byte) {
	a.neighbor.Clear()
	if hwAddr != nil {
		a.ether.HwAddr = hwAddr
	}
	if ipAddr != "" {
		a.setIpAddr(ipAddr)
	}
}

func (a *TapWorker) OnIpAddr(addr string) {
	a.eventQueue <- NewEvent(EvTapIpAddr, addr)
}

func (a *TapWorker) open() error {
	a.close()
	device, err := network.NewTaper(a.pinCfg.Network, a.devCfg)
	if err != nil {
		a.out.Error("TapWorker.open: %s", err)
		return err
	}
	device.Up() // up device firstly
	libol.Go(func() {
		a.Read(device)
	})
	a.out.Info("TapWorker.open: >>> %s <<<", device.Name())
	a.device = device
	if a.listener.OnOpen != nil {
		_ = a.listener.OnOpen(a)
	}
	return nil
}

func (a *TapWorker) newEth(t uint16, dst []byte) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = dst
	eth.Src = a.ether.HwAddr
	return eth
}

func (a *TapWorker) OnArpAlive(dest []byte) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.onMiss(dest)
}

// process if ethernet destination is missed
func (a *TapWorker) onMiss(dest []byte) {
	a.out.Debug("TapWorker.onMiss: %v.", dest)
	eth := a.newEth(libol.EthArp, libol.EthAll)
	reply := libol.NewArp()
	reply.OpCode = libol.ArpRequest
	reply.SIpAddr = a.ether.IpAddr
	reply.TIpAddr = dest
	reply.SHwAddr = a.ether.HwAddr
	reply.THwAddr = libol.EthZero

	frame := libol.NewFrameMessage(0)
	frame.Append(eth.Encode())
	frame.Append(reply.Encode())
	a.out.Debug("TapWorker.onMiss: %x.", frame.Frame()[:64])
	if a.listener.ReadAt != nil {
		_ = a.listener.ReadAt(frame)
	}
}

func (a *TapWorker) onFrame(frame *libol.FrameMessage, data []byte) int {
	size := len(data)
	if a.IsTun() {
		iph, err := libol.NewIpv4FromFrame(data)
		if err != nil {
			a.out.Warn("TapWorker.onFrame: %s", err)
			return 0
		}
		dest := iph.Destination
		if a.listener.FindNext != nil {
			dest = a.listener.FindNext(dest)
		}
		neb := a.neighbor.GetByBytes(dest)
		if neb == nil {
			a.onMiss(dest)
			a.out.Debug("TapWorker.onFrame: onMiss neighbor %v", dest)
			return 0
		}
		eth := a.newEth(libol.EthIp4, neb.HwAddr)
		frame.Append(eth.Encode()) // insert ethernet header.
		size += eth.Len
	}
	frame.SetSize(size)
	return size
}

func (a *TapWorker) Read(device network.Taper) {
	for {
		frame := libol.NewFrameMessage(0)
		data := frame.Frame()
		if a.IsTun() {
			data = data[libol.EtherLen:]
		}
		if n, err := device.Read(data); err != nil {
			a.out.Error("TapWorker.Read: %s", err)
			break
		} else {
			if a.out.Has(libol.DEBUG) {
				a.out.Debug("TapWorker.Read: %x", data[:n])
			}
			if size := a.onFrame(frame, data[:n]); size == 0 {
				continue
			}
			if a.listener.ReadAt != nil {
				_ = a.listener.ReadAt(frame)
			}
		}
	}
	if !a.isStopped() {
		a.eventQueue <- NewEvent(EvTapReadErr, "from read")
	}
}

func (a *TapWorker) dispatch(ev *WorkerEvent) {
	a.out.Event("TapWorker.dispatch: %s", ev)
	switch ev.Type {
	case EvTapReadErr, EvTapOpenErr, EvTapReset:
		if err := a.open(); err != nil {
			time.Sleep(time.Second * 2)
			a.eventQueue <- NewEvent(EvTapOpenErr, err.Error())
		}
	case EvTapIpAddr:
		a.setAddr(ev.Reason, nil)
	}
}

func (a *TapWorker) Loop() {
	for {
		select {
		case <-a.done:
			return
		case d := <-a.writeQueue:
			_ = a.DoWrite(d)
		case ev := <-a.eventQueue:
			a.lock.Lock()
			a.dispatch(ev)
			a.lock.Unlock()
		}
	}
}

func (a *TapWorker) DoWrite(frame *libol.FrameMessage) error {
	data := frame.Frame()
	if a.out.Has(libol.DEBUG) {
		a.out.Debug("TapWorker.DoWrite: %x", data)
	}
	a.lock.Lock()
	if a.device == nil {
		a.lock.Unlock()
		return libol.NewErr("device is nil")
	}
	if a.device.IsTun() {
		// proxy arp request.
		if a.toArp(data) {
			a.lock.Unlock()
			return nil
		}
		eth, err := libol.NewEtherFromFrame(data)
		if err != nil {
			a.out.Error("TapWorker.DoWrite: %s", err)
			a.lock.Unlock()
			return nil
		}
		if eth.IsIP4() {
			data = data[14:]
		} else {
			a.out.Debug("TapWorker.DoWrite: 0x%04x not IPv4", eth.Type)
			a.lock.Unlock()
			return nil
		}
	}
	a.lock.Unlock()

	proto, _ := frame.Proto()
	udp := proto.Udp
	if udp != nil && udp.Source == 53 {
		a.snoopDNS(udp.Payload)
	}

	if _, err := a.device.Write(data); err != nil {
		a.out.Error("TapWorker.DoWrite: %s", err)
		return err
	}
	return nil
}

func (a *TapWorker) Write(frame *libol.FrameMessage) error {
	a.writeQueue <- frame
	return nil
}

// learn source from arp
func (a *TapWorker) toArp(data []byte) bool {
	a.out.Debug("TapWorker.toArp")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		a.out.Warn("TapWorker.toArp: %s", err)
		return false
	}
	if !eth.IsArp() {
		return false
	}
	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		a.out.Error("TapWorker.toArp: %s.", err)
		return false
	}
	if arp.IsIP4() {
		if !bytes.Equal(eth.Src, arp.SHwAddr) {
			a.out.Error("TapWorker.toArp: eth.dst not arp.shw %x.", arp.SIpAddr)
			return true
		}
		switch arp.OpCode {
		case libol.ArpRequest:
			if bytes.Equal(arp.TIpAddr, a.ether.IpAddr) {
				eth := a.newEth(libol.EthArp, arp.SHwAddr)
				rep := libol.NewArp()
				rep.OpCode = libol.ArpReply
				rep.SIpAddr = a.ether.IpAddr
				rep.TIpAddr = arp.SIpAddr
				rep.SHwAddr = a.ether.HwAddr
				rep.THwAddr = arp.SHwAddr
				frame := libol.NewFrameMessage(0)
				frame.Append(eth.Encode())
				frame.Append(rep.Encode())
				a.out.Event("TapWorker.toArp: reply %v on %x.", rep.SIpAddr, rep.SHwAddr)
				if a.listener.ReadAt != nil {
					_ = a.listener.ReadAt(frame)
				}
			}
		case libol.ArpReply:
			// TODO learn by request.
			if bytes.Equal(arp.THwAddr, a.ether.HwAddr) {
				a.neighbor.Add(&Neighbor{
					HwAddr:  arp.SHwAddr,
					IpAddr:  arp.SIpAddr,
					NewTime: time.Now().Unix(),
					Uptime:  time.Now().Unix(),
				})
				a.out.Event("TapWorker.toArp: recv %v on %x.", arp.SIpAddr, arp.SHwAddr)
			}
		default:
			a.out.Warn("TapWorker.toArp: not op %x.", arp.OpCode)
		}
	}
	return true
}

func (a *TapWorker) close() {
	a.out.Info("TapWorker.close")
	if a.device != nil {
		if a.listener.OnClose != nil {
			a.listener.OnClose(a)
		}
		_ = a.device.Close()
	}
}

func (a *TapWorker) Start() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.out.Info("TapWorker.Start")
	libol.Go(a.Loop)
	libol.Go(a.neighbor.Start)
}

func (a *TapWorker) isStopped() bool {
	return a.device == nil
}

func (a *TapWorker) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.out.Info("TapWorker.Stop")
	a.done <- true
	a.neighbor.Stop()
	a.close()
	a.device = nil
}

func (a *TapWorker) snoopDNS(data []byte) {
	msg := new(dns.Msg)
	err := msg.Unpack(data)
	if err != nil {
		a.out.Info("Failed to unpack DNS message: %v\n", err)
		return
	}

	for _, rr := range msg.Answer {
		if n, ok := rr.(*dns.A); ok {
			a.listener.OnDNS(n.Hdr.Name, n.A)
		}
	}
}
