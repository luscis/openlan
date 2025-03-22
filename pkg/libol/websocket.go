package libol

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

type wsConn struct {
	*websocket.Conn
}

func (ws *wsConn) RemoteAddr() net.Addr {
	req := ws.Request()
	if req == nil {
		return ws.RemoteAddr()
	}
	addr := req.RemoteAddr
	if ret, err := net.ResolveTCPAddr("tcp", addr); err == nil {
		return ret
	}
	return nil
}

type CertConfig struct {
	Key      string
	Crt      string
	RootCa   string
	Insecure bool
}

type WebConfig struct {
	Cert    *CertConfig
	Block   *BlockCrypt
	Timeout time.Duration // ns
	RdQus   int           // per frames
	WrQus   int           // per frames
}

// Server Implement

type WebServer struct {
	*SocketServerImpl
	webCfg   *WebConfig
	listener *http.Server
}

func NewWebServer(listen string, cfg *WebConfig) *WebServer {
	t := &WebServer{
		webCfg:           cfg,
		SocketServerImpl: NewSocketServer(listen),
	}
	t.WrQus = cfg.WrQus
	t.close = t.Close
	return t
}

func (t *WebServer) Listen() (err error) {
	if t.webCfg.Cert != nil {
		Info("WebServer.Listen: wss://%s", t.address)
	} else {
		Info("WebServer.Listen: ws://%s", t.address)
	}
	t.listener = &http.Server{
		Addr: t.address,
	}
	return nil
}

func (t *WebServer) Close() {
	if t.listener != nil {
		_ = t.listener.Close()
		Info("WebServer.Close: %s", t.address)
		t.listener = nil
	}
}

func (t *WebServer) Accept() {
	Debug("WebServer.Accept")

	_ = t.Listen()
	defer t.Close()
	t.listener.Handler = websocket.Handler(func(ws *websocket.Conn) {
		if t.preAccept(ws, nil) != nil {
			return
		}
		defer ws.Close()
		ws.PayloadType = websocket.BinaryFrame
		wws := &wsConn{ws}
		client := NewWebClientFromConn(wws, t.webCfg)
		t.onClients <- client
		<-client.done
		Info("WebServer.Accept: %s exit", ws.RemoteAddr())
	})
	promise := Promise{
		First:  2 * time.Second,
		MinInt: 5 * time.Second,
		MaxInt: 30 * time.Second,
	}
	promise.Do(func() error {
		if t.webCfg.Cert == nil {
			if err := t.listener.ListenAndServe(); err != nil {
				Error("WebServer.Accept on %s: %s", t.address, err)
				return err
			}
		} else {
			ca := t.webCfg.Cert
			if err := t.listener.ListenAndServeTLS(ca.Crt, ca.Key); err != nil {
				Error("WebServer.Accept on %s: %s", t.address, err)
				return err
			}
		}
		return nil
	})
}

// Client Implement

type WebClient struct {
	*SocketClientImpl
	webCfg *WebConfig
	done   chan bool
	RdBuf  int // per frames
	WrBuf  int // per frames
}

func NewWebClient(addr string, cfg *WebConfig) *WebClient {
	t := &WebClient{
		webCfg: cfg,
		SocketClientImpl: NewSocketClient(SocketConfig{
			Address: addr,
			Block:   cfg.Block,
		}, &StreamMessagerImpl{
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
		done: make(chan bool, 2),
	}
	return t
}

func NewWebClientFromConn(conn net.Conn, cfg *WebConfig) *WebClient {
	addr := conn.RemoteAddr().String()
	t := &WebClient{
		webCfg: cfg,
		SocketClientImpl: NewSocketClient(SocketConfig{
			Address: addr,
			Block:   cfg.Block,
		}, &StreamMessagerImpl{
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
		done: make(chan bool, 2),
	}
	t.update(conn)
	return t
}

func (t *WebClient) GetCertPool(ca string) *x509.CertPool {
	caCert, err := os.ReadFile(ca)
	if err != nil {
		Error("WebClient.GetCertPool: %s", err)
		return nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		Warn("WebClient.GetCertPool: invalid cert")
	}
	return pool
}

func (t *WebClient) Connect() error {
	if !t.Retry() {
		return nil
	}
	var err error
	var config *websocket.Config
	if t.webCfg.Cert != nil {
		t.out.Info("WebClient.Connect: wss://%s", t.address)
		url := "wss://" + t.address
		if config, err = websocket.NewConfig(url, url); err != nil {
			return err
		}
		config.TlsConfig = &tls.Config{
			InsecureSkipVerify: t.webCfg.Cert.Insecure,
			RootCAs:            t.GetCertPool(t.webCfg.Cert.RootCa),
		}
	} else {
		t.out.Info("WebClient.Connect: ws://%s", t.address)
		url := "ws://" + t.address
		if config, err = websocket.NewConfig(url, url); err != nil {
			return err
		}
	}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	t.Reset(conn)
	if t.listener.OnConnected != nil {
		_ = t.listener.OnConnected(t)
	}
	return nil
}

func (t *WebClient) Close() {
	t.out.Debug("WebClient.Close: %v", t.IsOk())
	t.lock.Lock()
	if t.connection != nil {
		if t.status != ClTerminal {
			t.status = ClClosed
		}
		t.update(nil)
		t.done <- true
		t.private = nil
		t.lock.Unlock()
		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
		t.out.Debug("WebClient.Close: %d", t.status)
	} else {
		t.lock.Unlock()
	}
}

func (t *WebClient) Terminal() {
	t.SetStatus(ClTerminal)
	t.Close()
}

func (t *WebClient) SetStatus(v SocketStatus) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != v {
		if t.listener.OnStatus != nil {
			t.listener.OnStatus(t, t.status, v)
		}
		t.status = v
	}
}
