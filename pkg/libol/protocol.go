package libol

import (
	"encoding/binary"
	"fmt"
)

var (
	EthZero = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	EthAll  = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

const (
	EthArp  = 0x0806
	EthIp4  = 0x0800
	EthIp6  = 0x86DD
	EthVlan = 0x8100
)

type Ether struct {
	Dst  []byte
	Src  []byte
	Type uint16
	Len  int
}

const (
	EtherLen = 14
	VlanLen  = 4
	TcpLen   = 20
	Ipv4Len  = 20
	UdpLen   = 8
)

func NewEther(t uint16) (e *Ether) {
	e = &Ether{
		Type: t,
		Src:  make([]byte, 6),
		Dst:  make([]byte, 6),
		Len:  EtherLen,
	}
	return
}

func NewEtherArp() (e *Ether) {
	return NewEther(EthArp)
}

func NewEtherIP4() (e *Ether) {
	return NewEther(EthIp4)
}

func NewEtherFromFrame(frame []byte) (e *Ether, err error) {
	e = NewEther(0)
	err = e.Decode(frame)
	return
}

func (e *Ether) Decode(frame []byte) error {
	if len(frame) < 14 {
		return NewErr("Ether.Decode too small header: %d", len(frame))
	}

	copy(e.Dst[:6], frame[:6])
	copy(e.Src[:6], frame[6:12])
	e.Type = binary.BigEndian.Uint16(frame[12:14])
	e.Len = 14

	return nil
}

func (e *Ether) Encode() []byte {
	buffer := make([]byte, 14)

	copy(buffer[:6], e.Dst[:6])
	copy(buffer[6:12], e.Src[:6])
	binary.BigEndian.PutUint16(buffer[12:14], e.Type)

	return buffer[:14]
}

func (e *Ether) IsVlan() bool {
	return e.Type == EthVlan
}

func (e *Ether) IsArp() bool {
	return e.Type == EthArp
}

func (e *Ether) IsIP4() bool {
	return e.Type == EthIp4
}

type Vlan struct {
	Tci uint16
	Vid uint16
	Pro uint16
	Len int
}

func NewVlan(tci uint16, vid uint16) (n *Vlan) {
	n = &Vlan{
		Tci: tci,
		Vid: vid,
		Len: VlanLen,
	}

	return
}

func NewVlanFromFrame(frame []byte) (n *Vlan, err error) {
	n = &Vlan{
		Len: VlanLen,
	}
	err = n.Decode(frame)
	return
}

func (n *Vlan) Decode(frame []byte) error {
	if len(frame) < VlanLen {
		return NewErr("Vlan.Decode: too small header")
	}

	v := binary.BigEndian.Uint16(frame[0:2])
	n.Tci = uint16(v >> 12)
	n.Vid = uint16(0x0fff & v)
	n.Pro = binary.BigEndian.Uint16(frame[2:4])

	return nil
}

func (n *Vlan) Encode() []byte {
	buffer := make([]byte, 16)

	v := (n.Tci << 12) | n.Vid
	binary.BigEndian.PutUint16(buffer[0:2], v)
	binary.BigEndian.PutUint16(buffer[2:4], n.Pro)

	return buffer[:4]
}

const (
	ArpRequest = 1
	ArpReply   = 2
)

const (
	ArpHrdNetrom = 0
	ArpHrdEther  = 1
)

type Arp struct {
	HrdCode uint16 // format hardware address
	ProCode uint16 // format protocol address
	HrdLen  uint8  // length of hardware address
	ProLen  uint8  // length of protocol address
	OpCode  uint16 // ARP Op(command)

	SHwAddr []byte // sender hardware address.
	SIpAddr []byte // sender IP address.
	THwAddr []byte // target hardware address.
	TIpAddr []byte // target IP address.
	Len     int
}

func NewArp() (a *Arp) {
	a = &Arp{
		HrdCode: ArpHrdEther,
		ProCode: EthIp4,
		HrdLen:  6,
		ProLen:  4,
		OpCode:  ArpRequest,
		Len:     0,
		SHwAddr: make([]byte, 6),
		SIpAddr: make([]byte, 4),
		THwAddr: make([]byte, 6),
		TIpAddr: make([]byte, 4),
	}

	return
}

func NewArpFromFrame(frame []byte) (a *Arp, err error) {
	a = NewArp()
	err = a.Decode(frame)
	return
}

func (a *Arp) Decode(frame []byte) error {
	var err error

	if len(frame) < 8 {
		return NewErr("Arp.Decode: too small header: %d", len(frame))
	}

	a.HrdCode = binary.BigEndian.Uint16(frame[0:2])
	a.ProCode = binary.BigEndian.Uint16(frame[2:4])
	a.HrdLen = uint8(frame[4])
	a.ProLen = uint8(frame[5])
	if a.HrdLen != 6 || a.ProLen != 4 {
		return NewErr("Arp.Decode: AddrLen: %d,%d", a.HrdLen, a.ProLen)
	}
	a.OpCode = binary.BigEndian.Uint16(frame[6:8])

	p := uint8(8)
	if len(frame) < int(p+2*(a.HrdLen+a.ProLen)) {
		return NewErr("Arp.Decode: too small frame: %d", len(frame))
	}

	copy(a.SHwAddr[:6], frame[p:p+6])
	p += a.HrdLen
	copy(a.SIpAddr[:4], frame[p:p+4])
	p += a.ProLen
	copy(a.THwAddr[:6], frame[p:p+6])
	p += a.HrdLen
	copy(a.TIpAddr[:4], frame[p:p+4])
	p += a.ProLen

	a.Len = int(p)

	return err
}

func (a *Arp) Encode() []byte {
	buffer := make([]byte, 1024)

	binary.BigEndian.PutUint16(buffer[0:2], a.HrdCode)
	binary.BigEndian.PutUint16(buffer[2:4], a.ProCode)
	buffer[4] = byte(a.HrdLen)
	buffer[5] = byte(a.ProLen)
	binary.BigEndian.PutUint16(buffer[6:8], a.OpCode)

	p := uint8(8)
	copy(buffer[p:p+a.HrdLen], a.SHwAddr[0:a.HrdLen])
	p += a.HrdLen
	copy(buffer[p:p+a.ProLen], a.SIpAddr[0:a.ProLen])
	p += a.ProLen

	copy(buffer[p:p+a.HrdLen], a.THwAddr[0:a.HrdLen])
	p += a.HrdLen
	copy(buffer[p:p+a.ProLen], a.TIpAddr[0:a.ProLen])
	p += a.ProLen

	a.Len = int(p)

	return buffer[:p]
}

func (a *Arp) IsIP4() bool {
	return a.ProCode == EthIp4
}

func (a *Arp) IsReply() bool {
	return a.OpCode == ArpReply
}

func (a *Arp) IsRequest() bool {
	return a.OpCode == ArpRequest
}

const (
	Ipv4Ver = 0x04
	Ipv6Ver = 0x06
)

const (
	IpIcmp = 0x01
	IpIgmp = 0x02
	IpIpIp = 0x04
	IpTcp  = 0x06
	IpUdp  = 0x11
	IpEsp  = 0x32
	IpAh   = 0x33
	IpOspf = 0x59
	IpPim  = 0x67
	IpVrrp = 0x70
	IpIsis = 0x7c
)

func IpProto2Str(proto uint8) string {
	switch proto {
	case IpIcmp:
		return "icmp"
	case IpIgmp:
		return "igmp"
	case IpIpIp:
		return "ipip"
	case IpEsp:
		return "esp"
	case IpAh:
		return "ah"
	case IpOspf:
		return "ospf"
	case IpIsis:
		return "isis"
	case IpUdp:
		return "udp"
	case IpTcp:
		return "tcp"
	case IpPim:
		return "pim"
	case IpVrrp:
		return "vrrp"
	default:
		return fmt.Sprintf("%02x", proto)
	}
}

type Ipv4 struct {
	Version        uint8 //4bite v4: 0100, v6: 0110
	HeaderLen      uint8 //4bit 15*4
	ToS            uint8 //Type of Service
	TotalLen       uint16
	Identifier     uint16
	Flag           uint16 //3bit Z|DF|MF
	Offset         uint16 //13bit Fragment offset
	ToL            uint8  //Time of Live
	Protocol       uint8
	HeaderChecksum uint16 //Header Checksum
	Source         []byte
	Destination    []byte
	Options        uint32 //Reserved
	Len            int
}

func NewIpv4() (i *Ipv4) {
	i = &Ipv4{
		Version:        0x04,
		HeaderLen:      0x05,
		ToS:            0,
		TotalLen:       0,
		Identifier:     0,
		Flag:           0,
		Offset:         0,
		ToL:            0xff,
		Protocol:       0,
		HeaderChecksum: 0,
		Options:        0,
		Len:            Ipv4Len,
		Source:         make([]byte, 4),
		Destination:    make([]byte, 4),
	}
	return
}

func NewIpv4FromFrame(frame []byte) (i *Ipv4, err error) {
	i = NewIpv4()
	err = i.Decode(frame)
	return
}

func (i *Ipv4) Decode(frame []byte) error {
	if len(frame) < Ipv4Len {
		return NewErr("Ipv4.Decode: too small header: %d", len(frame))
	}

	h := uint8(frame[0])
	i.Version = h >> 4
	i.HeaderLen = h & 0x0f
	i.ToS = uint8(frame[1])
	i.TotalLen = binary.BigEndian.Uint16(frame[2:4])
	i.Identifier = binary.BigEndian.Uint16(frame[4:6])
	f := binary.BigEndian.Uint16(frame[6:8])
	i.Offset = f & 0x1fFf
	i.Flag = f >> 13
	i.ToL = uint8(frame[8])
	i.Protocol = uint8(frame[9])
	i.HeaderChecksum = binary.BigEndian.Uint16(frame[10:12])
	if !i.IsIP4() {
		return NewErr("Ipv4.Decode: not right ipv4 version: 0x%x", i.Version)
	}
	copy(i.Source[:4], frame[12:16])
	copy(i.Destination[:4], frame[16:20])

	return nil
}

func (i *Ipv4) Encode() []byte {
	buffer := make([]byte, 32)

	h := uint8((i.Version << 4) | i.HeaderLen)
	buffer[0] = h
	buffer[1] = i.ToS
	binary.BigEndian.PutUint16(buffer[2:4], i.TotalLen)
	binary.BigEndian.PutUint16(buffer[4:6], i.Identifier)
	f := uint16((i.Flag << 13) | i.Offset)
	binary.BigEndian.PutUint16(buffer[6:8], f)
	buffer[8] = i.ToL
	buffer[9] = i.Protocol
	binary.BigEndian.PutUint16(buffer[10:12], i.HeaderChecksum)
	copy(buffer[12:16], i.Source[:4])
	copy(buffer[16:20], i.Destination[:4])

	return buffer[:i.Len]
}

func (i *Ipv4) IsIP4() bool {
	return i.Version == Ipv4Ver
}

const (
	TcpUrg = 0x20
	TcpAck = 0x10
	TcpPsh = 0x08
	TcpRst = 0x04
	TcpSyn = 0x02
	TcpFin = 0x01
)

type Tcp struct {
	Source         uint16
	Destination    uint16
	Sequence       uint32
	Acknowledgment uint32
	DataOffset     uint8
	ControlBits    uint8
	Window         uint16
	Checksum       uint16
	UrgentPointer  uint16
	Options        []byte
	Padding        []byte
	Len            int
}

func NewTcp() (t *Tcp) {
	t = &Tcp{
		Source:         0,
		Destination:    0,
		Sequence:       0,
		Acknowledgment: 0,
		DataOffset:     0,
		ControlBits:    0,
		Window:         0,
		Checksum:       0,
		UrgentPointer:  0,
		Len:            TcpLen,
	}
	return
}

func NewTcpFromFrame(frame []byte) (t *Tcp, err error) {
	t = NewTcp()
	err = t.Decode(frame)
	return
}

func (t *Tcp) Decode(frame []byte) error {
	if len(frame) < TcpLen {
		return NewErr("Tcp.Decode: too small header: %d", len(frame))
	}

	t.Source = binary.BigEndian.Uint16(frame[0:2])
	t.Destination = binary.BigEndian.Uint16(frame[2:4])
	t.Sequence = binary.BigEndian.Uint32(frame[4:8])
	t.Acknowledgment = binary.BigEndian.Uint32(frame[8:12])
	t.DataOffset = uint8(frame[12])
	t.ControlBits = uint8(frame[13])
	t.Window = binary.BigEndian.Uint16(frame[14:16])
	t.Checksum = binary.BigEndian.Uint16(frame[16:18])
	t.UrgentPointer = binary.BigEndian.Uint16(frame[18:20])

	return nil
}

func (t *Tcp) Encode() []byte {
	buffer := make([]byte, 32)

	binary.BigEndian.PutUint16(buffer[0:2], t.Source)
	binary.BigEndian.PutUint16(buffer[2:4], t.Destination)
	binary.BigEndian.PutUint32(buffer[4:8], t.Sequence)
	binary.BigEndian.PutUint32(buffer[8:12], t.Acknowledgment)
	buffer[12] = t.DataOffset
	buffer[13] = t.ControlBits
	binary.BigEndian.PutUint16(buffer[14:16], t.Window)
	binary.BigEndian.PutUint16(buffer[16:18], t.Checksum)
	binary.BigEndian.PutUint16(buffer[18:20], t.UrgentPointer)

	return buffer[:t.Len]
}

func (t *Tcp) HasFlag(flag uint8) bool {
	return t.ControlBits&flag == flag
}

type Udp struct {
	Source      uint16
	Destination uint16
	Length      uint16
	Checksum    uint16
	Len         int
}

func NewUdp() (u *Udp) {
	u = &Udp{
		Source:      0,
		Destination: 0,
		Length:      0,
		Checksum:    0,
		Len:         UdpLen,
	}
	return
}

func NewUdpFromFrame(frame []byte) (u *Udp, err error) {
	u = NewUdp()
	err = u.Decode(frame)
	return
}

func (u *Udp) Decode(frame []byte) error {
	if len(frame) < UdpLen {
		return NewErr("Udp.Decode: too small header: %d", len(frame))
	}

	u.Source = binary.BigEndian.Uint16(frame[0:2])
	u.Destination = binary.BigEndian.Uint16(frame[2:4])
	u.Length = binary.BigEndian.Uint16(frame[4:6])
	u.Checksum = binary.BigEndian.Uint16(frame[6:8])

	return nil
}

func (u *Udp) Encode() []byte {
	buffer := make([]byte, 32)

	binary.BigEndian.PutUint16(buffer[0:2], u.Source)
	binary.BigEndian.PutUint16(buffer[2:4], u.Destination)
	binary.BigEndian.PutUint16(buffer[4:6], u.Length)
	binary.BigEndian.PutUint16(buffer[6:8], u.Checksum)

	return buffer[:u.Len]
}
