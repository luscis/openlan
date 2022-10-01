package libol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

const (
	MaxFrame = 1600
	MaxBuf   = 4096
	HlMI     = 0x02
	HlLI     = 0x04
	HlSize   = 0x04
	EthDI    = 0x06
	MaxMsg   = 1600 * 8
)

var MAGIC = []byte{0xff, 0xff}

const (
	LoginReq     = "logi= "
	LoginResp    = "logi: "
	NeighborReq  = "neig= "
	NeighborResp = "neig: "
	IpAddrReq    = "ipad= "
	IpAddrResp   = "ipad: "
	LeftReq      = "left= "
	SignReq      = "sign= "
	PingReq      = "ping= "
	PongResp     = "pong: "
	NegoReq      = "nego= "
	NegoResp     = "nego: "
)

func isControl(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	if bytes.Equal(data[:EthDI], EthZero[:EthDI]) {
		return true
	}
	return false
}

type FrameProto struct {
	// public
	Eth   *Ether
	Vlan  *Vlan
	Arp   *Arp
	Ip4   *Ipv4
	Udp   *Udp
	Tcp   *Tcp
	Err   error
	Frame []byte
}

func (i *FrameProto) Decode() error {
	data := i.Frame
	if i.Eth, i.Err = NewEtherFromFrame(data); i.Err != nil {
		return i.Err
	}
	data = data[i.Eth.Len:]
	if i.Eth.IsVlan() {
		if i.Vlan, i.Err = NewVlanFromFrame(data); i.Err != nil {
			return i.Err
		}
		data = data[i.Vlan.Len:]
	}
	switch i.Eth.Type {
	case EthIp4:
		if i.Ip4, i.Err = NewIpv4FromFrame(data); i.Err != nil {
			return i.Err
		}
		data = data[i.Ip4.Len:]
		switch i.Ip4.Protocol {
		case IpTcp:
			if i.Tcp, i.Err = NewTcpFromFrame(data); i.Err != nil {
				return i.Err
			}
		case IpUdp:
			if i.Udp, i.Err = NewUdpFromFrame(data); i.Err != nil {
				return i.Err
			}
		}
	case EthArp:
		if i.Arp, i.Err = NewArpFromFrame(data); i.Err != nil {
			return i.Err
		}
	}
	return nil
}

type FrameMessage struct {
	seq     uint64
	control bool
	action  string
	params  []byte
	buffer  []byte
	size    int
	total   int
	frame   []byte
	proto   *FrameProto
}

func NewFrameMessage(maxSize int) *FrameMessage {
	if maxSize <= 0 {
		maxSize = MaxBuf
	}
	maxSize += HlSize + EthDI
	if HasLog(DEBUG) {
		Debug("NewFrameMessage: size %d", maxSize)
	}
	m := FrameMessage{
		params: make([]byte, 0, 2),
		buffer: make([]byte, maxSize),
	}
	m.frame = m.buffer[HlSize:]
	m.total = len(m.frame)
	return &m
}

func NewFrameMessageFromBytes(buffer []byte) *FrameMessage {
	m := FrameMessage{
		params: make([]byte, 0, 2),
		buffer: buffer,
	}
	m.frame = m.buffer[HlSize:]
	m.total = len(m.frame)
	m.size = len(m.frame)
	m.control = isControl(m.frame)
	if m.control {
		m.Decode()
	}
	return &m
}

func (m *FrameMessage) Decode() bool {
	if m.control {
		if len(m.frame) < 2*EthDI {
			Warn("FrameMessage.Decode: too small message")
		} else {
			m.action = string(m.frame[EthDI : 2*EthDI])
			m.params = m.frame[2*EthDI:]
		}
	}
	return m.control
}

func (m *FrameMessage) IsEthernet() bool {
	return !m.control
}

func (m *FrameMessage) IsControl() bool {
	return m.control
}

func (m *FrameMessage) Frame() []byte {
	return m.frame
}

func (m *FrameMessage) String() string {
	return fmt.Sprintf("control: %t, frame: %x", m.control, m.frame[:20])
}

func (m *FrameMessage) Action() string {
	return m.action
}

func (m *FrameMessage) CmdAndParams() (string, []byte) {
	return m.action, m.params
}

func (m *FrameMessage) Append(data []byte) {
	add := len(data)
	if m.total-m.size >= add {
		copy(m.frame[m.size:], data)
		m.size += add
	} else {
		Warn("FrameMessage.Append: %d not enough buffer", m.total)
	}
}

func (m *FrameMessage) Size() int {
	return m.size
}

func (m *FrameMessage) SetSize(v int) {
	m.size = v
}

func (m *FrameMessage) Proto() (*FrameProto, error) {
	if m.proto == nil {
		m.proto = &FrameProto{Frame: m.frame}
		_ = m.proto.Decode()
	}
	return m.proto, m.proto.Err
}

type ControlMessage struct {
	seq      uint64
	control  bool
	operator string
	action   string
	params   []byte
}

func NewControlFrame(action string, body []byte) *FrameMessage {
	m := NewControlMessage(action[:4], action[4:], body)
	return m.Encode()
}

//operator: request is '= ', and response is  ': '
//action: login, network etc.
//body: json string.
func NewControlMessage(action, opr string, body []byte) *ControlMessage {
	c := ControlMessage{
		control:  true,
		action:   action,
		params:   body,
		operator: opr,
	}
	return &c
}

func (c *ControlMessage) Encode() *FrameMessage {
	p := fmt.Sprintf("%s%s%s", c.action[:4], c.operator[:2], c.params)
	frame := NewFrameMessage(len(p))
	frame.control = c.control
	frame.action = c.action + c.operator
	frame.params = c.params
	frame.Append(EthZero[:6])
	frame.Append([]byte(p))
	return frame
}

type Messager interface {
	Crypt() *BlockCrypt
	SetCrypt(*BlockCrypt)
	Send(conn net.Conn, frame *FrameMessage) (int, error)
	Receive(conn net.Conn, max, min int) (*FrameMessage, error)
	Flush()
}

type StreamMessagerImpl struct {
	timeout time.Duration // ns for read and write deadline.
	block   *BlockCrypt
	buffer  []byte
	bufSize int // default is (1518 + 20+20+14) * 8
}

func (s *StreamMessagerImpl) SetCrypt(block *BlockCrypt) {
	s.block = CopyBlockCrypt(block)
}

func (s *StreamMessagerImpl) Crypt() *BlockCrypt {
	return s.block
}

func (s *StreamMessagerImpl) Flush() {
	s.buffer = nil
}

func (s *StreamMessagerImpl) write(conn net.Conn, tmp []byte) (int, error) {
	if s.timeout != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	n, err := conn.Write(tmp)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *StreamMessagerImpl) writeX(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	size := len(buf)
	left := size - offset
	if HasLog(LOG) {
		Log("StreamMessagerImpl.writeX: %s %d", conn.RemoteAddr(), size)
		Log("StreamMessagerImpl.writeX: %s Data %x", conn.RemoteAddr(), buf)
	}
	for left > 0 {
		tmp := buf[offset:]
		if HasLog(LOG) {
			Log("StreamMessagerImpl.writeX: tmp %s %d", conn.RemoteAddr(), len(tmp))
		}
		n, err := s.write(conn, tmp)
		if err != nil {
			return err
		}
		if HasLog(LOG) {
			Log("StreamMessagerImpl.writeX: %s snd %d, size %d", conn.RemoteAddr(), n, size)
		}
		offset += n
		left = size - offset
	}
	return nil
}

func (s *StreamMessagerImpl) encode(frame *FrameMessage) {
	frame.buffer[0] = MAGIC[0]
	frame.buffer[1] = MAGIC[1]
	binary.BigEndian.PutUint16(frame.buffer[HlMI:HlLI], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
}

func (s *StreamMessagerImpl) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	s.encode(frame)
	fs := frame.size + HlSize
	if err := s.writeX(conn, frame.buffer[:fs]); err != nil {
		return 0, err
	}
	return fs, nil
}

func (s *StreamMessagerImpl) read(conn net.Conn, tmp []byte) (int, error) {
	if s.timeout != 0 {
		err := conn.SetReadDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	n, err := conn.Read(tmp)
	if err != nil {
		return 0, err
	}
	return n, nil
}

//340Mib
func (s *StreamMessagerImpl) readX(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	left := len(buf)
	if HasLog(LOG) {
		Log("StreamMessagerImpl.readX: %s %d", conn.RemoteAddr(), len(buf))
	}
	for left > 0 {
		tmp := make([]byte, left)
		n, err := s.read(conn, tmp)
		if err != nil {
			return err
		}
		copy(buf[offset:], tmp)
		offset += n
		left -= n
	}
	if HasLog(LOG) {
		Log("StreamMessagerImpl.readX: Data %s %x", conn.RemoteAddr(), buf)
	}
	return nil
}

func (s *StreamMessagerImpl) decode(tmp []byte, min int) (*FrameMessage, error) {
	ts := len(tmp)
	if ts < min {
		return nil, nil
	}
	if !bytes.Equal(tmp[:HlMI], MAGIC[:HlMI]) {
		return nil, NewErr("wrong magic")
	}
	ps := binary.BigEndian.Uint16(tmp[HlMI:HlLI])
	fs := int(ps) + HlSize
	if ts >= fs {
		s.buffer = tmp[fs:]
		if s.block != nil {
			s.block.Decrypt(tmp[HlSize:fs], tmp[HlSize:fs])
		}
		if HasLog(DEBUG) {
			Debug("StreamMessagerImpl.decode: %d %x", fs, tmp[:fs])
		}
		return NewFrameMessageFromBytes(tmp[:fs]), nil
	}
	return nil, nil
}

// 430Mib
func (s *StreamMessagerImpl) Receive(conn net.Conn, max, min int) (*FrameMessage, error) {
	frame, err := s.decode(s.buffer, min)
	if err != nil {
		return nil, err
	}
	if frame != nil { // firstly, check buffer has messages.
		return frame, nil
	}
	if s.bufSize == 0 {
		s.bufSize = MaxMsg // 1572 * 8
	}
	bs := len(s.buffer)
	tmp := make([]byte, s.bufSize)
	if bs > 0 {
		copy(tmp[:bs], s.buffer[:bs])
	}
	for { // loop forever until socket error or find one message.
		rn, err := s.read(conn, tmp[bs:])
		if err != nil {
			return nil, err
		}
		rs := bs + rn
		frame, err := s.decode(tmp[:rs], min)
		if err != nil {
			return nil, err
		}
		if frame != nil {
			return frame, nil
		}
		// If notFound message, continue to read.
		bs = rs
	}
}

type PacketMessagerImpl struct {
	timeout time.Duration // ns for read and write deadline
	block   *BlockCrypt
	bufSize int // default is (1518 + 20+20+14) * 8
}

func (s *PacketMessagerImpl) SetCrypt(block *BlockCrypt) {
	s.block = CopyBlockCrypt(block)
}

func (s *PacketMessagerImpl) Crypt() *BlockCrypt {
	return s.block
}

func (s *PacketMessagerImpl) Flush() {
	//TODO
}

func (s *PacketMessagerImpl) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	frame.buffer[0] = MAGIC[0]
	frame.buffer[1] = MAGIC[1]
	binary.BigEndian.PutUint16(frame.buffer[HlMI:HlLI], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
	if HasLog(DEBUG) {
		Debug("PacketMessagerImpl.Send: %s %d %x", conn.RemoteAddr(), frame.size, frame.buffer)
	}
	if s.timeout != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	if _, err := conn.Write(frame.buffer[:HlSize+frame.size]); err != nil {
		return 0, err
	}
	return frame.size, nil
}

func (s *PacketMessagerImpl) Receive(conn net.Conn, max, min int) (*FrameMessage, error) {
	if s.bufSize == 0 {
		s.bufSize = MaxMsg
	}
	frame := NewFrameMessage(s.bufSize)
	if HasLog(DEBUG) {
		Debug("PacketMessagerImpl.Receive %s %d", conn.RemoteAddr(), s.timeout)
	}
	if s.timeout != 0 {
		err := conn.SetReadDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return nil, err
		}
	}
	n, err := conn.Read(frame.buffer)
	if err != nil {
		return nil, err
	}
	if HasLog(DEBUG) {
		Debug("PacketMessagerImpl.Receive: %s %x", conn.RemoteAddr(), frame.buffer[:n])
	}
	if n <= 4 {
		return nil, NewErr("%s: small frame", conn.RemoteAddr())
	}
	if !bytes.Equal(frame.buffer[:HlMI], MAGIC[:HlMI]) {
		return nil, NewErr("%s: wrong magic", conn.RemoteAddr())
	}
	size := int(binary.BigEndian.Uint16(frame.buffer[HlMI:HlLI]))
	if size > max || size < min {
		return nil, NewErr("%s: wrong size %d", conn.RemoteAddr(), size)
	}
	tmp := frame.buffer[HlSize : HlSize+size]
	if s.block != nil {
		s.block.Decrypt(tmp, tmp)
	}
	frame.size = size
	frame.frame = tmp
	return frame, nil
}

type BlockCrypt struct {
	kcp.BlockCrypt
	algorithm string
	key       string
}

func GetKcpBlock(algo string, key string) kcp.BlockCrypt {
	var block kcp.BlockCrypt

	pass := make([]byte, 64)
	if len(key) <= 64 {
		copy(pass, key)
	} else {
		copy(pass, key[:64])
	}

	switch algo {
	case "aes-128":
		block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-256":
		block, _ = kcp.NewAESBlockCrypt(pass[:32])
	case "xor":
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	default:
		block, _ = kcp.NewNoneBlockCrypt(pass)
	}

	return block
}

func NewBlockCrypt(algo string, key string) *BlockCrypt {
	if key == "" {
		return nil
	}
	return &BlockCrypt{
		BlockCrypt: GetKcpBlock(algo, key),
		algorithm:  algo,
		key:        key,
	}
}

func CopyBlockCrypt(crypt *BlockCrypt) *BlockCrypt {
	if crypt == nil {
		return nil
	}
	return &BlockCrypt{
		BlockCrypt: GetKcpBlock(crypt.algorithm, crypt.key),
		algorithm:  crypt.algorithm,
		key:        crypt.key,
	}
}

func (b *BlockCrypt) Update(key string) {
	b.key = key
	b.BlockCrypt = GetKcpBlock(b.algorithm, b.key)
}
