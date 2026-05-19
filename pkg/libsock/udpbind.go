package libsock

import (
	"net"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/libol"
)

type UDPBind struct {
	lock       sync.RWMutex
	bufSize    int
	connection *net.UDPConn
	address    *net.UDPAddr
	sessions   *libol.SafeStrMap
	accept     chan *UDPBindConn
}

func UDPBindListen(addr string, clients, bufSize int) (net.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	if bufSize == 0 {
		bufSize = 4096
	}
	libol.Debug("bufSize: %d", bufSize)
	x := &UDPBind{
		address:  udpAddr,
		sessions: libol.NewSafeStrMap(clients),
		accept:   make(chan *UDPBindConn, 2),
		bufSize:  bufSize,
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	x.connection = conn
	libol.Go(x.Loop)
	return x, nil
}

func (x *UDPBind) Recv(udpAddr *net.UDPAddr, data []byte) error {
	// dispatch to UDPBindConn and new accept
	addr := udpAddr.String()
	if obj, ok := x.sessions.GetEx(addr); ok {
		conn := obj.(*UDPBindConn)
		conn.toQueue(data)
		return nil
	}
	conn := &UDPBindConn{
		connection: x.connection,
		remoteAddr: udpAddr,
		localAddr:  x.address,
		readQueue:  make(chan []byte, 1024),
		closed:     false,
		onClose: func(conn *UDPBindConn) {
			libol.Info("UDPBind.Recv: onClose %s", conn)
			x.sessions.Del(addr)
		},
	}
	if err := x.sessions.Set(addr, conn); err != nil {
		return libol.NewErr("session.Set: %s", err)
	}
	x.accept <- conn
	conn.toQueue(data)
	return nil
}

// Loop forever
func (x *UDPBind) Loop() {
	for {
		data := make([]byte, x.bufSize)
		n, udpAddr, err := x.connection.ReadFromUDP(data)
		if err != nil {
			libol.Error("UDPBind.Loop %s", err)
			break
		}
		if err := x.Recv(udpAddr, data[:n]); err != nil {
			libol.Warn("UDPBind.Loop: %s", err)
		}
	}
}

// Accept waits for and returns the next connection to the listener.
func (x *UDPBind) Accept() (net.Conn, error) {
	return <-x.accept, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (x *UDPBind) Close() error {
	x.lock.Lock()
	defer x.lock.Unlock()

	_ = x.connection.Close()
	return nil
}

// returns the listener's network address.
func (x *UDPBind) Addr() net.Addr {
	return x.address
}

type UDPBindConn struct {
	lock       sync.RWMutex
	connection *net.UDPConn
	remoteAddr *net.UDPAddr
	localAddr  *net.UDPAddr
	readQueue  chan []byte
	closed     bool
	readDead   time.Time
	writeDead  time.Time
	onClose    func(conn *UDPBindConn)
}

func (c *UDPBindConn) toQueue(b []byte) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return
	} else {
		c.lock.RUnlock()
	}
	c.readQueue <- b
}

func (c *UDPBindConn) Read(b []byte) (n int, err error) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return 0, libol.NewErr("read on closed")
	}
	var timeout *time.Timer
	outChan := make(<-chan time.Time)
	if !c.readDead.IsZero() {
		if time.Now().After(c.readDead) {
			c.lock.RUnlock()
			return 0, libol.NewErr("read timeout")
		}
		delay := time.Until(c.readDead)
		timeout = time.NewTimer(delay)
		outChan = timeout.C
	}
	c.lock.RUnlock()

	// wait for read event or timeout or error
	select {
	case <-outChan:
		return 0, libol.NewErr("read timeout")
	case d := <-c.readQueue:
		if timeout != nil {
			timeout.Stop()
		}
		return copy(b, d), nil
	}
}

func (c *UDPBindConn) Write(b []byte) (n int, err error) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return 0, libol.NewErr("write to closed")
	} else {
		c.lock.RUnlock()
	}
	return c.connection.WriteToUDP(b, c.remoteAddr)
}

func (c *UDPBindConn) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return nil
	}
	if c.onClose != nil {
		c.onClose(c)
	}
	c.connection = nil
	c.closed = true

	return nil
}

func (c *UDPBindConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *UDPBindConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *UDPBindConn) SetDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.readDead = t
	c.writeDead = t
	return nil
}

func (c *UDPBindConn) SetReadDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.readDead = t
	return nil
}

func (c *UDPBindConn) SetWriteDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.writeDead = t
	return nil
}

func (c *UDPBindConn) String() string {
	return c.remoteAddr.String()
}
