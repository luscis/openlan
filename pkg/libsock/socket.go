package libsock

import (
	"bytes"
	"crypto/md5"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luscis/openlan/pkg/libol"
)

const (
	ClInit        = 0x00
	ClConnected   = 0x01
	ClUnAuth      = 0x02
	ClAuth        = 0x03
	ClConnecting  = 0x04
	ClTerminal    = 0x05
	ClClosed      = 0x06
	ClNegotiating = 0x07
	ClNegotiated  = 0x08
)

type SocketStatus uint8

func (s SocketStatus) String() string {
	switch s {
	case ClInit:
		return "initialized"
	case ClConnected:
		return "connected"
	case ClUnAuth:
		return "unauthenticated"
	case ClAuth:
		return "authenticated"
	case ClClosed:
		return "closed"
	case ClConnecting:
		return "connecting"
	case ClTerminal:
		return "terminal"
	case ClNegotiating:
		return "negotiating"
	case ClNegotiated:
		return "negotiated"
	}
	return ""
}

//  Socket Client Interface and Implement

const (
	CsSendOkay  = "send"
	CsRecvOkay  = "receive"
	CsSendError = "error"
	CsDropped   = "dropped"
)

const (
	CryptLevelGlobal  = "global"
	CryptLevelNetwork = "network"
)

type ClientCryptDecl struct {
	Network string
	Level   string
}

type ClientListener struct {
	OnClose     func(client SocketClient) error
	OnConnected func(client SocketClient) error
	OnStatus    func(client SocketClient, old, new SocketStatus)
}

type SocketClient interface {
	LocalAddr() string
	RemoteAddr() string
	Protocol() string
	Connect() error
	Close()
	WriteMsg(frame *FrameMessage) error
	ReadMsg() (*FrameMessage, error)
	UpTime() int64
	AliveTime() int64
	String() string
	Terminal()
	Private() any
	SetPrivate(v any)
	Status() SocketStatus
	SetStatus(v SocketStatus)
	MinSize() int
	IsOk() bool
	Have(status SocketStatus) bool
	Statistics() map[string]int64
	SetListener(listener ClientListener)
	SetTimeout(v int64)
	Out() *libol.SubLogger
	SetKey(key string)
	Key() string
}

type StreamSocket struct {
	message    Messager
	connection net.Conn
	minSize    int
	protocol   string
	out        *libol.SubLogger
	remoteAddr string
	localAddr  string
	address    string
	Block      *BlockCrypt
	sendOkay   atomic.Int64
	recvOkay   atomic.Int64
	sendError  atomic.Int64
	dropped    atomic.Int64
}

func (t *StreamSocket) LocalAddr() string {
	return t.localAddr
}

func (t *StreamSocket) RemoteAddr() string {
	return t.remoteAddr
}

func (t *StreamSocket) Protocol() string {
	return t.protocol
}

func (t *StreamSocket) String() string {
	return t.protocol + ":" + t.remoteAddr
}

func (t *StreamSocket) IsOk() bool {
	return t.connection != nil
}

func (t *StreamSocket) WriteMsg(frame *FrameMessage) error {
	if !t.IsOk() {
		t.dropped.Add(1)
		return libol.NewErr("%s not okay", t)
	}
	if libol.HasLog(libol.CMD) && frame.IsControl() {
		action, params := frame.CmdAndParams()
		libol.Cmd("StreamSocket.WriteMsg: %s%s", action, params)
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessagerImpl{}
	}
	size, err := t.message.Send(t.connection, frame)
	if err != nil {
		t.sendError.Add(1)
		return err
	}
	t.sendOkay.Add(int64(size))
	return nil
}

func (t *StreamSocket) ReadMsg() (*FrameMessage, error) {
	if libol.HasLog(libol.LOG) {
		libol.Log("StreamSocket.ReadMsg: %s", t)
	}
	if !t.IsOk() {
		return nil, libol.NewErr("%s not okay", t)
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessagerImpl{}
	}
	frame, err := t.message.Receive(t.connection, t.minSize)
	if err != nil {
		return nil, err
	}
	size := len(frame.frame)
	t.recvOkay.Add(int64(size))
	return frame, nil
}

func (t *StreamSocket) SetKey(key string) {
	if block := t.message.Crypt(); block != nil {
		block.Update(key)
	}
}

func (t *StreamSocket) Key() string {
	key := ""
	if block := t.message.Crypt(); block != nil {
		key = block.key
	}
	return key
}

type SocketConfig struct {
	Address  string
	Protocol string
	Block    *BlockCrypt
}

type SocketClientImpl struct {
	*StreamSocket
	lock          sync.RWMutex
	listener      ClientListener
	newTime       int64
	connectedTime int64
	private       any
	status        SocketStatus
	timeout       int64 // sec for read and write timeout
}

func NewSocketClient(cfg SocketConfig, message Messager) *SocketClientImpl {
	client := &SocketClientImpl{
		StreamSocket: &StreamSocket{
			minSize:    15,
			message:    message,
			protocol:   cfg.Protocol,
			remoteAddr: cfg.Address,
			address:    cfg.Address,
			Block:      cfg.Block,
		},
		newTime: time.Now().Unix(),
		status:  ClInit,
	}
	client.out = libol.NewSubLogger(client.String())
	return client
}

func (s *SocketClientImpl) Negotiate() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.Key() == "" {
		libol.Warn("SocketClientImpl.negotiate: skip, empty pre-shared key")
		return nil
	}
	key := string(libol.GenLetters(64))
	request := NewControlFrame(NegoReq, []byte(key))
	magic := "ffff"
	network := ""
	if s.private != nil {
		if decl, ok := s.private.(ClientCryptDecl); ok {
			if strings.EqualFold(decl.Level, CryptLevelNetwork) {
				setFrameMagic(request, MAGICv1, decl.Network)
				magic = "v1"
				network = decl.Network
			}
		}
	}
	libol.Info("SocketClientImpl.negotiate: send request magic=%s network=%s", magic, network)
	if err := s.WriteMsg(request); err != nil {
		libol.Error("SocketClientImpl.negotiate: write request failed: %s", err)
		return err
	}
	s.status = ClNegotiating
	reply, err := s.ReadMsg()
	if err != nil {
		return err
	}
	if !reply.IsControl() {
		libol.Info("SocketClientImpl.negotiate %s", reply.String())
		return libol.NewErr("wrong message type")
	}
	action, params := reply.CmdAndParams()
	if action != NegoResp {
		return libol.NewErr("wrong message type: %s", action)
	}
	libol.Cmd("SocketClientImpl.negotiate %s %x", action, params)
	sum := md5.Sum([]byte(key))
	if !bytes.Equal(sum[:md5.Size], params) {
		return libol.NewErr("negotiate key failed: %x != %x", key, params)
	}
	if block := s.message.Crypt(); block != nil {
		block.Update(string(key))
	}
	s.status = ClNegotiated
	return nil
}

// MUST IMPLEMENT
func (s *SocketClientImpl) Connect() error {
	return nil
}

// MUST IMPLEMENT
func (s *SocketClientImpl) Close() {
}

// MUST IMPLEMENT
func (s *SocketClientImpl) Terminal() {
}

func (s *SocketClientImpl) Out() *libol.SubLogger {
	if s.out == nil {
		s.out = libol.NewSubLogger(s.address)
	}
	return s.out
}

func (s *SocketClientImpl) Retry() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.connection != nil ||
		s.status == ClTerminal ||
		s.status == ClUnAuth {
		return false
	}
	s.status = ClConnecting
	return true
}

func (s *SocketClientImpl) Status() SocketStatus {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status
}

func (s *SocketClientImpl) UpTime() int64 {
	return time.Now().Unix() - s.newTime
}

func (s *SocketClientImpl) AliveTime() int64 {
	if s.connectedTime == 0 {
		return 0
	}
	return time.Now().Unix() - s.connectedTime
}

func (s *SocketClientImpl) Private() any {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.private
}

func (s *SocketClientImpl) SetPrivate(v any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.private = v
}

func (s *SocketClientImpl) MinSize() int {
	return s.minSize
}

func (s *SocketClientImpl) Have(state SocketStatus) bool {
	return s.Status() == state
}

func (s *SocketClientImpl) Statistics() map[string]int64 {
	return map[string]int64{
		CsSendOkay:  s.sendOkay.Load(),
		CsRecvOkay:  s.recvOkay.Load(),
		CsSendError: s.sendError.Load(),
		CsDropped:   s.dropped.Load(),
	}
}

func (s *SocketClientImpl) SetListener(listener ClientListener) {
	s.listener = listener
}

func (s *SocketClientImpl) SetTimeout(v int64) {
	s.timeout = v
}

func (s *SocketClientImpl) update(conn net.Conn) {
	if conn != nil {
		s.connection = conn
		s.connectedTime = time.Now().Unix()
		s.localAddr = conn.LocalAddr().String()
		s.remoteAddr = conn.RemoteAddr().String()
	} else {
		if s.connection != nil {
			_ = s.connection.Close()
		}
		s.connection = nil
		s.localAddr = ""
		s.remoteAddr = ""
		s.message.Flush()
	}
	if s.Block != nil {
		s.message.SetCrypt(s.Block)
	}
	libol.Info("SocketClientImpl.update: %s %s", s.localAddr, s.remoteAddr)
}

func (s *SocketClientImpl) Try(conn net.Conn) {
	s.lock.Lock()
	s.update(conn)
	s.status = ClConnected
	s.lock.Unlock()
	libol.Info("SocketClientImpl.Try: to connected")
	if err := s.Negotiate(); err != nil {
		libol.Error("SocketClientImpl.Try %s", err)
		return
	}
}

// MUST IMPLEMENT
func (s *SocketClientImpl) SetStatus(v SocketStatus) {
}

// Socket Server Interface and Implement

const (
	SsRecv   = "receive"
	SsDeny   = "deny"
	SsAlive  = "alive"
	SsSend   = "send"
	SsDrop   = "dropped"
	SsAccept = "accept"
	SsClose  = "closed"
)

type ServerListener struct {
	OnClient func(client SocketClient) error
	OnClose  func(client SocketClient) error
	ReadAt   func(client SocketClient, f *FrameMessage) error
}

type ReadClient func(client SocketClient, f *FrameMessage) error

type SocketServer interface {
	Listen() (err error)
	Close()
	Accept()
	ListClient() <-chan SocketClient
	UpdateCrypt(block *BlockCrypt)
	OffClient(client SocketClient)
	TotalClient() int
	Loop(call ServerListener)
	Read(client SocketClient, ReadAt ReadClient)
	String() string
	Address() string
	Statistics() map[string]int64
	SetTimeout(v int64)
	Protocol() string
}

// TODO keepalive to release zombie connections.
type SocketServerImpl struct {
	lock       sync.RWMutex
	statistics *libol.SafeStrInt64
	address    string
	maxClient  int
	clients    *libol.SafeStrMap
	onClients  chan SocketClient
	offClients chan SocketClient
	close      func()
	timeout    int64 // sec for read and write timeout
	WrQus      int   // per frames.
	error      error
}

const negotiateReadTimeout = 10 * time.Second

func NewSocketServer(listen string) *SocketServerImpl {
	return &SocketServerImpl{
		address:    listen,
		statistics: libol.NewSafeStrInt64(),
		maxClient:  128,
		clients:    libol.NewSafeStrMap(1024),
		onClients:  make(chan SocketClient, 1024),
		offClients: make(chan SocketClient, 1024),
		WrQus:      1024,
	}
}

func (t *SocketServerImpl) Protocol() string {
	return "unknown"
}

func (t *SocketServerImpl) ListClient() <-chan SocketClient {
	list := make(chan SocketClient, 32)
	libol.Go(func() {
		t.clients.Iter(func(k string, v any) {
			if client, ok := v.(SocketClient); ok {
				list <- client
			}
		})
		list <- nil
	})
	return list
}

func (t *SocketServerImpl) TotalClient() int {
	return t.clients.Len()
}

func (t *SocketServerImpl) OffClient(client SocketClient) {
	libol.Warn("SocketServerImpl.OffClient %s", client)
	if client != nil {
		t.offClients <- client
	}
}

func (t *SocketServerImpl) UpdateCrypt(block *BlockCrypt) {
	// implemented by concrete servers when they have protocol config to update
}

func (t *SocketServerImpl) kickAllClients() {
	for client := range t.ListClient() {
		if client == nil {
			break
		}
		t.OffClient(client)
	}
}

func (t *SocketServerImpl) Negotiate(client SocketClient) error {
	if client.Key() == "" {
		return nil
	}
	if sc, ok := client.(*SocketClientImpl); ok && sc.connection != nil {
		_ = sc.connection.SetReadDeadline(time.Now().Add(negotiateReadTimeout))
		defer func() {
			_ = sc.connection.SetReadDeadline(time.Time{})
		}()
	}
	libol.Info("SocketServerImpl.Negotiate: waiting request from %s (timeout %s)", client, negotiateReadTimeout)
	request, err := client.ReadMsg()
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			libol.Warn("SocketServerImpl.Negotiate: timeout waiting request from %s", client)
		}
		return err
	}
	if !request.IsControl() {
		libol.Warn("SocketServerImpl.Negotiate: except control but %s", request.String())
		return libol.NewErr("wrong message type")
	}
	libol.Info("SocketServerImpl.Negotiate: request received from %s %v %s", client, request.magic, request.network)
	client.SetStatus(ClNegotiated)
	action, params := request.CmdAndParams()
	if action == NegoReq {
		key := params
		if len(key) == 0 {
			return libol.NewErr("empty negotiate key")
		}
		sum := md5.Sum(key)
		libol.Cmd("SocketServerImpl.Negotiate: key:%s", key)
		reply := NewControlFrame(NegoResp, sum[:md5.Size])
		if err := client.WriteMsg(reply); err != nil {
			return err
		}
		client.SetKey(string(key))
		return nil
	}
	return libol.NewErr("wrong message type: %s", action)

}

func (t *SocketServerImpl) doOnClient(call ServerListener, client SocketClient) {
	libol.Info("SocketServerImpl.doOnClient: %s", client)
	_ = t.clients.Set(client.RemoteAddr(), client)
	if call.OnClient != nil {
		libol.Go(func() {
			if err := t.Negotiate(client); err != nil {
				t.OffClient(client)
				libol.Warn("SocketServerImpl.doOnClient: %s %s", client, err)
				return
			}
			_ = call.OnClient(client)
			if call.ReadAt != nil {
				t.Read(client, call.ReadAt)
			}
		})
	}
}

func (t *SocketServerImpl) doOffClient(call ServerListener, client SocketClient) {
	libol.Info("SocketServerImpl.doOffClient: %s", client)
	addr := client.RemoteAddr()
	if _, ok := t.clients.GetEx(addr); ok {
		libol.Info("SocketServerImpl.doOffClient: close %s", addr)
		t.statistics.Add(SsClose, 1)
		if call.OnClose != nil {
			_ = call.OnClose(client)
		}
		client.Close()
		t.clients.Del(addr)
		t.statistics.Add(SsAlive, -1)
	}
}

func (t *SocketServerImpl) Loop(call ServerListener) {
	libol.Debug("SocketServerImpl.Loop")
	defer t.close()
	for {
		select {
		case client := <-t.onClients:
			t.doOnClient(call, client)
		case client := <-t.offClients:
			t.doOffClient(call, client)
		}
	}
}

func (t *SocketServerImpl) Read(client SocketClient, ReadAt ReadClient) {
	libol.Log("SocketServerImpl.Read: %s", client)
	done := make(chan bool, 2)
	queue := make(chan *FrameMessage, t.WrQus)
	libol.Go(func() {
		for {
			select {
			case frame := <-queue:
				if err := ReadAt(client, frame); err != nil {
					libol.Error("SocketServerImpl.Read: readAt %s", err)
					return
				}
			case <-done:
				return
			}
		}
	})
	for {
		frame, err := client.ReadMsg()
		if err != nil || frame.size <= 0 {
			if frame != nil {
				libol.Error("SocketServerImpl.Read: %s %d", client, frame.size)
			} else {
				libol.Error("SocketServerImpl.Read: %s %s", client, err)
			}
			done <- true
			t.OffClient(client)
			break
		}
		t.statistics.Add(SsRecv, 1)
		if libol.HasLog(libol.LOG) {
			libol.Log("SocketServerImpl.Read: length: %d ", frame.size)
			libol.Log("SocketServerImpl.Read: frame : %x", frame)
		}
		queue <- frame
	}
}

// MUST IMPLEMENT
func (t *SocketServerImpl) Listen() error {
	return nil
}

// MUST IMPLEMENT
func (t *SocketServerImpl) Accept() {
}

// MUST IMPLEMENT
func (t *SocketServerImpl) Close() {
	if t.close != nil {
		t.close()
	}
}

func (t *SocketServerImpl) Address() string {
	return t.address
}

func (t *SocketServerImpl) String() string {
	return t.Address()
}

func (t *SocketServerImpl) Statistics() map[string]int64 {
	sts := make(map[string]int64, 32)
	t.statistics.Copy(sts)
	return sts
}

func (t *SocketServerImpl) SetTimeout(v int64) {
	t.timeout = v
}

// pre-process when accept connection,
// and allowed accept new connection, will return nil.
func (t *SocketServerImpl) preAccept(conn net.Conn, err error) error {
	if err != nil {
		if t.error == nil || t.error.Error() != err.Error() {
			libol.Warn("SocketServerImpl.preAccept: %s", err)
		}
		t.error = err
		return err
	}
	t.error = nil
	addr := conn.RemoteAddr()
	libol.Debug("SocketServerImpl.preAccept: %s", addr)
	t.statistics.Add(SsAccept, 1)
	alive := t.statistics.Get(SsAlive)
	if alive >= int64(t.maxClient) {
		libol.Debug("SocketServerImpl.preAccept: close %s", addr)
		t.statistics.Add(SsDeny, 1)
		t.statistics.Add(SsClose, 1)
		_ = conn.Close()
		return libol.NewErr("too many open clients")
	}
	libol.Debug("SocketServerImpl.preAccept: allow %s", addr)
	t.statistics.Add(SsAlive, 1)
	return nil
}
