package libol

import (
	"net"
	"sync"
	"time"
)

type XDP struct {
	lock       sync.RWMutex
	bufSize    int
	connection *net.UDPConn
	address    *net.UDPAddr
	sessions   *SafeStrMap
	accept     chan *XDPConn
}

func XDPListen(addr string, clients, bufSize int) (net.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	if bufSize == 0 {
		bufSize = MaxBuf
	}
	Debug("bufSize: %d", bufSize)
	x := &XDP{
		address:  udpAddr,
		sessions: NewSafeStrMap(clients),
		accept:   make(chan *XDPConn, 2),
		bufSize:  bufSize,
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	x.connection = conn
	Go(x.Loop)
	return x, nil
}

func (x *XDP) Recv(udpAddr *net.UDPAddr, data []byte) error {
	// dispatch to XDPConn and new accept
	addr := udpAddr.String()
	if obj, ok := x.sessions.GetEx(addr); ok {
		conn := obj.(*XDPConn)
		conn.toQueue(data)
		return nil
	}
	conn := &XDPConn{
		connection: x.connection,
		remoteAddr: udpAddr,
		localAddr:  x.address,
		readQueue:  make(chan []byte, 1024),
		closed:     false,
		onClose: func(conn *XDPConn) {
			Info("XDP.Recv: onClose %s", conn)
			x.sessions.Del(addr)
		},
	}
	if err := x.sessions.Set(addr, conn); err != nil {
		return NewErr("session.Set: %s", err)
	}
	x.accept <- conn
	conn.toQueue(data)
	return nil
}

// Loop forever
func (x *XDP) Loop() {
	for {
		data := make([]byte, x.bufSize)
		n, udpAddr, err := x.connection.ReadFromUDP(data)
		if err != nil {
			Error("XDP.Loop %s", err)
			break
		}
		if err := x.Recv(udpAddr, data[:n]); err != nil {
			Warn("XDP.Loop: %s", err)
		}
	}
}

// Accept waits for and returns the next connection to the listener.
func (x *XDP) Accept() (net.Conn, error) {
	return <-x.accept, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (x *XDP) Close() error {
	x.lock.Lock()
	defer x.lock.Unlock()

	_ = x.connection.Close()
	return nil
}

// returns the listener's network address.
func (x *XDP) Addr() net.Addr {
	return x.address
}

type XDPConn struct {
	lock       sync.RWMutex
	connection *net.UDPConn
	remoteAddr *net.UDPAddr
	localAddr  *net.UDPAddr
	readQueue  chan []byte
	closed     bool
	readDead   time.Time
	writeDead  time.Time
	onClose    func(conn *XDPConn)
}

func (c *XDPConn) toQueue(b []byte) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return
	} else {
		c.lock.RUnlock()
	}
	c.readQueue <- b
}

func (c *XDPConn) Read(b []byte) (n int, err error) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return 0, NewErr("read on closed")
	}
	var timeout *time.Timer
	outChan := make(<-chan time.Time)
	if !c.readDead.IsZero() {
		if time.Now().After(c.readDead) {
			c.lock.RUnlock()
			return 0, NewErr("read timeout")
		}
		delay := c.readDead.Sub(time.Now())
		timeout = time.NewTimer(delay)
		outChan = timeout.C
	}
	c.lock.RUnlock()

	// wait for read event or timeout or error
	select {
	case <-outChan:
		return 0, NewErr("read timeout")
	case d := <-c.readQueue:
		if timeout != nil {
			timeout.Stop()
		}
		return copy(b, d), nil
	}
}

func (c *XDPConn) Write(b []byte) (n int, err error) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return 0, NewErr("write to closed")
	} else {
		c.lock.RUnlock()
	}
	return c.connection.WriteToUDP(b, c.remoteAddr)
}

func (c *XDPConn) Close() error {
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

func (c *XDPConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *XDPConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *XDPConn) SetDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.readDead = t
	c.writeDead = t
	return nil
}

func (c *XDPConn) SetReadDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.readDead = t
	return nil
}

func (c *XDPConn) SetWriteDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.writeDead = t
	return nil
}

func (c *XDPConn) String() string {
	return c.remoteAddr.String()
}
