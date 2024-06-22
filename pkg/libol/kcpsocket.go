package libol

import (
	"net"
	"time"

	"github.com/xtaci/kcp-go/v5"
)

type KcpConfig struct {
	Block        *BlockCrypt
	WinSize      int           // default 1024
	DataShards   int           // default 10
	ParityShards int           // default 3
	Timeout      time.Duration // ns
	RdQus        int           // per frames
	WrQus        int           // per frames
}

var defaultKcpConfig = KcpConfig{
	Block:        nil,
	WinSize:      1024,
	DataShards:   10,
	ParityShards: 3,
	Timeout:      120 * time.Second,
}

func NewKcpConfig() *KcpConfig {
	return &defaultKcpConfig
}

type KcpServer struct {
	*SocketServerImpl
	kcpCfg   *KcpConfig
	listener *kcp.Listener
}

func setConn(conn *kcp.UDPSession, cfg *KcpConfig) {
	Info("setConn %s", conn.RemoteAddr())
	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	Info("setConn %s to fast3", conn.RemoteAddr())
	// normal: 0, 40, 2, 1
	// fast  : 0, 30, 2, 1
	// fast3 : 1, 10, 2, 1
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(cfg.WinSize, cfg.WinSize)
	conn.SetACKNoDelay(true)
}

func NewKcpServer(listen string, cfg *KcpConfig) *KcpServer {
	if cfg == nil {
		cfg = &defaultKcpConfig
	}
	k := &KcpServer{
		kcpCfg:           cfg,
		SocketServerImpl: NewSocketServer(listen),
	}
	k.close = k.Close
	return k
}

func (k *KcpServer) Listen() (err error) {
	k.listener, err = kcp.ListenWithOptions(
		k.address,
		nil,
		k.kcpCfg.DataShards,
		k.kcpCfg.ParityShards)
	if err != nil {
		k.listener = nil
		return err
	}
	if err := k.listener.SetDSCP(46); err != nil {
		Warn("KcpServer.SetDSCP %s", err)
	}
	Info("KcpServer.Listen: kcp://%s", k.address)
	return nil
}

func (k *KcpServer) Close() {
	if k.listener != nil {
		_ = k.listener.Close()
		Info("KcpServer.Close: %s", k.address)
		k.listener = nil
	}
}

func (k *KcpServer) Accept() {
	Debug("KcpServer.Accept")
	promise := Promise{
		First:  2 * time.Second,
		MinInt: 5 * time.Second,
		MaxInt: 30 * time.Second,
	}
	promise.Do(func() error {
		if err := k.Listen(); err != nil {
			Warn("KcpServer.Accept: %s", err)
			return err
		}
		return nil
	})
	defer k.Close()
	for {
		if k.listener == nil {
			return
		}
		conn, err := k.listener.AcceptKCP()
		if k.preAccept(conn, err) != nil {
			continue
		}
		setConn(conn, k.kcpCfg)
		k.onClients <- NewKcpClientFromConn(conn, k.kcpCfg)
	}
}

// Client Implement

type KcpClient struct {
	*SocketClientImpl
	kcpCfg *KcpConfig
}

func NewKcpClient(addr string, cfg *KcpConfig) *KcpClient {
	if cfg == nil {
		cfg = &defaultKcpConfig
	}
	c := &KcpClient{
		kcpCfg: cfg,
		SocketClientImpl: NewSocketClient(SocketConfig{
			Address: addr,
			Block:   cfg.Block,
		}, &StreamMessagerImpl{
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
	}
	return c
}

func NewKcpClientFromConn(conn net.Conn, cfg *KcpConfig) *KcpClient {
	if cfg == nil {
		cfg = &defaultKcpConfig
	}
	addr := conn.RemoteAddr().String()
	c := &KcpClient{
		SocketClientImpl: NewSocketClient(SocketConfig{
			Address: addr,
			Block:   cfg.Block,
		}, &StreamMessagerImpl{
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
	}
	c.update(conn)
	return c
}

func (c *KcpClient) Connect() error {
	if !c.Retry() {
		return nil
	}
	c.out.Info("KcpClient.Connect: kcp://%s", c.address)
	conn, err := kcp.DialWithOptions(
		c.address,
		nil,
		c.kcpCfg.DataShards,
		c.kcpCfg.DataShards)
	if err != nil {
		return err
	}
	if err := conn.SetDSCP(46); err != nil {
		c.out.Warn("KcpClient.SetDSCP: ", err)
	}
	setConn(conn, c.kcpCfg)
	c.Reset(conn)
	if c.listener.OnConnected != nil {
		_ = c.listener.OnConnected(c)
	}
	return nil
}

func (c *KcpClient) Close() {
	c.out.Debug("KcpClient.Close: %v", c.IsOk())
	c.lock.Lock()
	if c.connection != nil {
		if c.status != ClTerminal {
			c.status = ClClosed
		}
		c.out.Debug("KcpClient.Close")
		c.update(nil)
		c.private = nil
		c.lock.Unlock()
		if c.listener.OnClose != nil {
			_ = c.listener.OnClose(c)
		}
	} else {
		c.lock.Unlock()
	}
}

func (c *KcpClient) Terminal() {
	c.SetStatus(ClTerminal)
	c.Close()
}

func (c *KcpClient) SetStatus(v SocketStatus) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != v {
		if c.listener.OnStatus != nil {
			c.listener.OnStatus(c, c.status, v)
		}
		c.status = v
	}
}
