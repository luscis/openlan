package libol

import (
	"crypto/tls"
	"net"
	"time"
)

type TcpConfig struct {
	Tls     *tls.Config
	Block   *BlockCrypt
	Timeout time.Duration // ns
	RdQus   int           // per frames
	WrQus   int           // per frames
}

// Server Implement

type TcpServer struct {
	*SocketServerImpl
	tcpCfg   *TcpConfig
	listener net.Listener
}

func NewTcpServer(listen string, cfg *TcpConfig) *TcpServer {
	t := &TcpServer{
		tcpCfg:           cfg,
		SocketServerImpl: NewSocketServer(listen),
	}
	t.WrQus = cfg.WrQus
	t.close = t.Close
	return t
}

func (t *TcpServer) Listen() (err error) {
	if t.tcpCfg.Tls != nil {
		t.listener, err = tls.Listen("tcp", t.address, t.tcpCfg.Tls)
		if err != nil {
			t.listener = nil
			return err
		}
		Info("TcpServer.Listen: tls://%s", t.address)
	} else {
		t.listener, err = net.Listen("tcp", t.address)
		if err != nil {
			t.listener = nil
			return err
		}
		Info("TcpServer.Listen: tcp://%s", t.address)
	}
	return nil
}

func (t *TcpServer) Close() {
	if t.listener != nil {
		_ = t.listener.Close()
		Info("TcpServer.Close: %s", t.address)
		t.listener = nil
	}
}

func (t *TcpServer) Accept() {
	Debug("TcpServer.Accept")
	promise := Promise{
		First:  2 * time.Second,
		MinInt: 5 * time.Second,
		MaxInt: 30 * time.Second,
	}
	promise.Do(func() error {
		if err := t.Listen(); err != nil {
			Warn("TcpServer.Accept: %s", err)
			return err
		}
		return nil
	})
	defer t.Close()
	for {
		if t.listener == nil {
			return
		}
		conn, err := t.listener.Accept()
		if t.preAccept(conn, err) != nil {
			continue
		}
		t.onClients <- NewTcpClientFromConn(conn, t.tcpCfg)
	}
}

// Client Implement

type TcpClient struct {
	*SocketClientImpl
	tcpCfg *TcpConfig
}

func NewTcpClient(addr string, cfg *TcpConfig) *TcpClient {
	t := &TcpClient{
		tcpCfg: cfg,
		SocketClientImpl: NewSocketClient(SocketConfig{
			Address: addr,
			Block:   cfg.Block,
		}, &StreamMessagerImpl{
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
	}
	return t
}

func NewTcpClientFromConn(conn net.Conn, cfg *TcpConfig) *TcpClient {
	addr := conn.RemoteAddr().String()
	t := &TcpClient{
		tcpCfg: cfg,
		SocketClientImpl: NewSocketClient(SocketConfig{
			Address: addr,
			Block:   cfg.Block,
		}, &StreamMessagerImpl{
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
	}
	t.update(conn)
	return t
}

func (t *TcpClient) Connect() error {
	if !t.Retry() {
		return nil
	}
	var err error
	var conn net.Conn
	if t.tcpCfg.Tls != nil {
		t.out.Info("TcpClient.Connect: tls://%s", t.address)
		conn, err = tls.Dial("tcp", t.address, t.tcpCfg.Tls)
	} else {
		t.out.Info("TcpClient.Connect: tcp://%s", t.address)
		conn, err = net.Dial("tcp", t.address)
	}
	if err != nil {
		return err
	}
	t.Reset(conn)
	if t.listener.OnConnected != nil {
		_ = t.listener.OnConnected(t)
	}
	return nil
}

func (t *TcpClient) Close() {
	t.out.Debug("TcpClient.Close: %v", t.IsOk())
	t.lock.Lock()
	if t.connection != nil {
		if t.status != ClTerminal {
			t.status = ClClosed
		}
		t.update(nil)
		t.private = nil
		t.lock.Unlock()
		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
		t.out.Debug("TcpClient.Close: %d", t.status)
	} else {
		t.lock.Unlock()
	}
}

func (t *TcpClient) Terminal() {
	t.SetStatus(ClTerminal)
	t.Close()
}

func (t *TcpClient) SetStatus(v SocketStatus) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != v {
		if t.listener.OnStatus != nil {
			t.listener.OnStatus(t, t.status, v)
		}
		t.status = v
	}
}
