package libsock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/xtaci/kcp-go/v5"
)

const (
	MaxFrame = 1600
	MaxBuf   = 4096
	HMSize   = 0x02
	HlSize   = 0x04 // magic 2, size 2
	Hv1Size  = 0x13 // magic 2, size 2, networkv1 15
	EthDI    = 0x06
	MaxMsg   = 1600 * 8
	V1Size   = 15
)

var MAGIC = [2]byte{0xff, 0xff}
var MAGICv1 = [2]byte{0xff, 0x00}
var ResolveNetworkCrypt func(network string) *BlockCrypt

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
	if bytes.Equal(data[:EthDI], libol.EthZero[:EthDI]) {
		return true
	}
	return false
}

type FrameProto struct {
	// public
	Eth   *libol.Ether
	Vlan  *libol.Vlan
	Arp   *libol.Arp
	Ip4   *libol.Ipv4
	Udp   *libol.Udp
	Tcp   *libol.Tcp
	Err   error
	Frame []byte
}

func (i *FrameProto) Decode() error {
	data := i.Frame
	if i.Eth, i.Err = libol.NewEtherFromFrame(data); i.Err != nil {
		return i.Err
	}
	data = data[i.Eth.Len:]
	if i.Eth.IsVlan() {
		if i.Vlan, i.Err = libol.NewVlanFromFrame(data); i.Err != nil {
			return i.Err
		}
		data = data[i.Vlan.Len:]
	}
	switch i.Eth.Type {
	case libol.EthIp4:
		if i.Ip4, i.Err = libol.NewIpv4FromFrame(data); i.Err != nil {
			return i.Err
		}
		data = data[i.Ip4.Len:]
		switch i.Ip4.Protocol {
		case libol.IpTcp:
			if i.Tcp, i.Err = libol.NewTcpFromFrame(data); i.Err != nil {
				return i.Err
			}
		case libol.IpUdp:
			if i.Udp, i.Err = libol.NewUdpFromFrame(data); i.Err != nil {
				return i.Err
			}
		}
	case libol.EthArp:
		if i.Arp, i.Err = libol.NewArpFromFrame(data); i.Err != nil {
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
	magic   [2]byte
	network string
	buffer  []byte
	size    int // size of packet or control message
	total   int
	frame   []byte
	proto   *FrameProto
}

func NewFrameMessage(maxSize int) *FrameMessage {
	if maxSize <= 0 {
		maxSize = MaxBuf
	}
	maxSize += HlSize + EthDI
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("NewFrameMessage: size %d", maxSize)
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
	m.magic = [2]byte{buffer[0], buffer[1]}
	m.frame = m.buffer[HlSize:]
	m.total = len(m.frame)
	m.size = len(m.frame)
	m.control = isControl(m.frame)
	if m.control {
		m.Decode()
	}
	return &m
}

func (m *FrameMessage) MagicV1() bool {
	return m.magic == MAGICv1
}

func (m *FrameMessage) Magic() [2]byte {
	if m.magic == [2]byte{} {
		m.magic = MAGIC
	}
	return m.magic
}

func (m *FrameMessage) Decode() bool {
	if m.control {
		if len(m.frame) < 2*EthDI {
			libol.Warn("FrameMessage.Decode: too small message")
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
		libol.Warn("FrameMessage.Append: %d not enough buffer", m.total)
	}
}

func (m *FrameMessage) Size() int {
	return m.size
}

func (m *FrameMessage) SetSize(v int) {
	m.size = v
}

func setFrameMagic(m *FrameMessage, magic [2]byte, network string) {
	if m == nil {
		return
	}
	m.magic = magic
	m.network = network
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

// operator: request is '= ', and response is  ': '
// action: login, network etc.
// body: json string.
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
	frame.Append(libol.EthZero[:EthDI])
	frame.Append([]byte(p))
	return frame
}

type Messager interface {
	Crypt() *BlockCrypt
	SetCrypt(*BlockCrypt)
	Send(conn net.Conn, frame *FrameMessage) (int, error)
	Receive(conn net.Conn, min int) (*FrameMessage, error)
	Flush()
}

func encodeMagicV1Frame(network string, payload []byte) []byte {
	buf := make([]byte, len(payload)+Hv1Size)
	buf[0] = MAGICv1[0]
	buf[1] = MAGICv1[1]
	binary.BigEndian.PutUint16(buf[HMSize:HlSize], uint16(V1Size+len(payload)))
	copy(buf[HlSize:Hv1Size], []byte(network))
	copy(buf[Hv1Size:], payload)
	return buf
}

type frameHeader struct {
	magic       [2]byte
	headerLen   int
	frameAt     int
	payloadSize int // bytes after magic+len (for v1 includes network id)
	frameLen    int // pure frame bytes (without network id)
	network     string
}

func (h *frameHeader) MagicV1() bool {
	return h.magic == MAGICv1
}

func decodeFrameHeader(buf []byte, min int) (*frameHeader, error) {
	if len(buf) < min {
		return nil, nil
	}
	h := &frameHeader{
		magic:     [2]byte{buf[0], buf[1]},
		headerLen: HlSize,
		frameAt:   HlSize,
	}
	if h.magic == MAGICv1 {
		h.headerLen = Hv1Size
		h.frameAt = Hv1Size
	}
	if len(buf) < h.headerLen {
		return nil, nil
	}
	if h.magic == MAGICv1 {
		name := bytes.TrimRight(buf[HlSize:Hv1Size], "\x00")
		h.network = string(name)
		h.payloadSize = int(binary.BigEndian.Uint16(buf[HMSize:HlSize]))
		if h.payloadSize < V1Size {
			return nil, libol.NewErr("wrong size %d", h.payloadSize)
		}
		h.frameLen = h.payloadSize - V1Size
	} else {
		if h.magic != MAGIC {
			return nil, libol.NewErr("wrong magic: %x", h.magic)
		}
		h.payloadSize = int(binary.BigEndian.Uint16(buf[HMSize:HlSize]))
		h.frameLen = h.payloadSize
	}
	if h.frameLen < min {
		return nil, libol.NewErr("wrong size %d", h.frameLen)
	}
	return h, nil
}

type StreamMessagerImpl struct {
	timeout time.Duration // ns for read and write deadline.
	block   *BlockCrypt
	buffer  []byte
	readBuf []byte
	bufSize int // default is (1518 + 20+20+14) * 8
	readAt  time.Time
	writeAt time.Time
}

func (s *StreamMessagerImpl) SetCrypt(block *BlockCrypt) {
	s.block = CopyBlockCrypt(block)
}

func (s *StreamMessagerImpl) Crypt() *BlockCrypt {
	return s.block
}

func (s *StreamMessagerImpl) Flush() {
	s.buffer = nil
	s.readBuf = nil
	s.readAt = time.Time{}
	s.writeAt = time.Time{}
}

func shouldRefreshDeadline(last time.Time, timeout time.Duration, now time.Time) bool {
	if timeout == 0 {
		return false
	}
	if last.IsZero() || !now.Before(last) {
		return true
	}
	refresh := timeout / 4
	if refresh <= 0 {
		refresh = timeout
	}
	return last.Sub(now) <= refresh
}

func (s *StreamMessagerImpl) write(conn net.Conn, tmp []byte) (int, error) {
	now := time.Now()
	if shouldRefreshDeadline(s.writeAt, s.timeout, now) {
		s.writeAt = now.Add(s.timeout)
		err := conn.SetWriteDeadline(s.writeAt)
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
		return libol.NewErr("connection is nil")
	}
	size := len(buf)
	if libol.HasLog(libol.LOG) {
		libol.Log("StreamMessagerImpl.writeX: %d %s Data %x", size, conn.RemoteAddr(), buf)
	}
	n, err := s.write(conn, buf)
	if err != nil {
		return err
	}
	if n >= size {
		return nil
	}
	offset := n
	left := size - offset
	for left > 0 {
		tmp := buf[offset:]
		if libol.HasLog(libol.LOG) {
			libol.Log("StreamMessagerImpl.writeX: tmp %s %d", conn.RemoteAddr(), len(tmp))
		}
		n, err := s.write(conn, tmp)
		if err != nil {
			return err
		}
		if libol.HasLog(libol.LOG) {
			libol.Log("StreamMessagerImpl.writeX: %s snd %d, size %d", conn.RemoteAddr(), n, size)
		}
		offset += n
		left = size - offset
	}
	return nil
}

func (s *StreamMessagerImpl) encode(frame *FrameMessage) {
	binary.BigEndian.PutUint16(frame.buffer[HMSize:HlSize], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
}

func (s *StreamMessagerImpl) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	s.encode(frame)
	magic := frame.Magic()
	if frame.MagicV1() {
		network := frame.network
		buf := encodeMagicV1Frame(network, frame.frame[:frame.size])
		if err := s.writeX(conn, buf); err != nil {
			return 0, err
		}
		return len(buf), nil
	}
	fs := frame.size + HlSize
	frame.buffer[0] = magic[0]
	frame.buffer[1] = magic[1]
	if err := s.writeX(conn, frame.buffer[:fs]); err != nil {
		return 0, err
	}
	return fs, nil
}

func (s *StreamMessagerImpl) read(conn net.Conn, tmp []byte) (int, error) {
	now := time.Now()
	if shouldRefreshDeadline(s.readAt, s.timeout, now) {
		s.readAt = now.Add(s.timeout)
		err := conn.SetReadDeadline(s.readAt)
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

// 340Mib
func (s *StreamMessagerImpl) readX(conn net.Conn, buf []byte) error {
	if conn == nil {
		return libol.NewErr("connection is nil")
	}
	offset := 0
	left := len(buf)
	if libol.HasLog(libol.LOG) {
		libol.Log("StreamMessagerImpl.readX: %s %d", conn.RemoteAddr(), len(buf))
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
	if libol.HasLog(libol.LOG) {
		libol.Log("StreamMessagerImpl.readX: Data %s %x", conn.RemoteAddr(), buf)
	}
	return nil
}

func (s *StreamMessagerImpl) decode(tmp []byte, min int) (*FrameMessage, error) {
	ts := len(tmp)
	h, err := decodeFrameHeader(tmp, min)
	if err != nil || h == nil {
		return nil, err
	}
	if h.MagicV1() && ResolveNetworkCrypt != nil {
		if block := ResolveNetworkCrypt(h.network); block != nil {
			s.SetCrypt(block)
			libol.Info("StreamMessagerImpl.decode: resolved network %v", block)
		}
	}

	// Build the frame message.
	fs := HlSize + h.payloadSize
	if ts >= fs {
		s.buffer = tmp[fs:]
		frameData := tmp[h.frameAt : h.frameAt+h.frameLen]
		if s.block != nil {
			s.block.Decrypt(frameData, frameData)
		}
		if libol.HasLog(libol.DEBUG) {
			libol.Debug("StreamMessagerImpl.decode: %d %x", fs, tmp[:fs])
		}
		buf := make([]byte, HlSize+h.frameLen)
		copy(buf, h.magic[:])
		binary.BigEndian.PutUint16(buf[HMSize:HlSize], uint16(h.frameLen))
		copy(buf[HlSize:], frameData)

		newframe := NewFrameMessageFromBytes(buf)
		setFrameMagic(newframe, h.magic, h.network)
		return newframe, nil
	}
	return nil, nil
}

// 430Mib
func (s *StreamMessagerImpl) Receive(conn net.Conn, min int) (*FrameMessage, error) {
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
	if len(s.readBuf) != s.bufSize {
		s.readBuf = make([]byte, s.bufSize)
	}
	tmp := s.readBuf
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
		if rs >= len(tmp) {
			return nil, libol.NewErr("%s: frame too large or incomplete (%d bytes buffered)", conn.RemoteAddr(), rs)
		}
		// If notFound message, continue to read.
		bs = rs
	}
}

type PacketMessagerImpl struct {
	timeout time.Duration // ns for read and write deadline
	block   *BlockCrypt
	bufSize int // default is (1518 + 20+20+14) * 8
	readAt  time.Time
	writeAt time.Time
}

func (s *PacketMessagerImpl) SetCrypt(block *BlockCrypt) {
	s.block = CopyBlockCrypt(block)
}

func (s *PacketMessagerImpl) Crypt() *BlockCrypt {
	return s.block
}

func (s *PacketMessagerImpl) Flush() {
	s.readAt = time.Time{}
	s.writeAt = time.Time{}
}

func (s *PacketMessagerImpl) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	magic := frame.Magic()
	binary.BigEndian.PutUint16(frame.buffer[HMSize:HlSize], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("PacketMessagerImpl.Send: %s %d %x", conn.RemoteAddr(), frame.size, frame.buffer)
	}
	now := time.Now()
	if shouldRefreshDeadline(s.writeAt, s.timeout, now) {
		s.writeAt = now.Add(s.timeout)
		err := conn.SetWriteDeadline(s.writeAt)
		if err != nil {
			return 0, err
		}
	}
	if frame.MagicV1() {
		network := frame.network
		buf := encodeMagicV1Frame(network, frame.frame[:frame.size])
		if n, err := conn.Write(buf); err != nil {
			return 0, err
		} else {
			return n, nil
		}
	}
	sz := HlSize + frame.size
	frame.buffer[0] = magic[0]
	frame.buffer[1] = magic[1]
	if n, err := conn.Write(frame.buffer[:sz]); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

func (s *PacketMessagerImpl) Receive(conn net.Conn, min int) (*FrameMessage, error) {
	if s.bufSize == 0 {
		s.bufSize = MaxMsg
	}
	frame := NewFrameMessage(s.bufSize)
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("PacketMessagerImpl.Receive %s %d", conn.RemoteAddr(), s.timeout)
	}
	now := time.Now()
	if shouldRefreshDeadline(s.readAt, s.timeout, now) {
		s.readAt = now.Add(s.timeout)
		err := conn.SetReadDeadline(s.readAt)
		if err != nil {
			return nil, err
		}
	}
	n, err := conn.Read(frame.buffer)
	if err != nil {
		return nil, err
	}
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("PacketMessagerImpl.Receive: %s %x", conn.RemoteAddr(), frame.buffer[:n])
	}
	if n <= 4 {
		return nil, libol.NewErr("%s: small frame", conn.RemoteAddr())
	}
	h, err := decodeFrameHeader(frame.buffer[:n], min)
	if err != nil || h == nil {
		if err != nil {
			return nil, libol.NewErr("%s: %s", conn.RemoteAddr(), err)
		}
		return nil, libol.NewErr("%s: small frame", conn.RemoteAddr())
	}
	if h.MagicV1() && ResolveNetworkCrypt != nil {
		if block := ResolveNetworkCrypt(h.network); block != nil {
			s.SetCrypt(block)
		}
	}
	if HlSize+h.payloadSize > n {
		return nil, libol.NewErr("%s: truncated frame frameLen=%d recv=%d", conn.RemoteAddr(), h.frameLen, n)
	}

	// Build the frame message.
	frameData := frame.buffer[h.frameAt : h.frameAt+h.frameLen]
	if s.block != nil {
		s.block.Decrypt(frameData, frameData)
	}

	buf := make([]byte, HlSize+h.frameLen)
	buf[0] = MAGIC[0]
	buf[1] = MAGIC[1]
	binary.BigEndian.PutUint16(buf[HMSize:HlSize], uint16(h.frameLen))
	copy(buf[HlSize:], frameData)

	newFrame := NewFrameMessageFromBytes(buf)
	setFrameMagic(newFrame, h.magic, h.network)
	return newFrame, nil
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
	case "sm4":
		block, _ = kcp.NewSM4BlockCrypt(pass[:16])
	case "salsa20":
		block, _ = kcp.NewSalsa20BlockCrypt(pass[:32])
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
