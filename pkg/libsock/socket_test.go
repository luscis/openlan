package libsock

import (
	"io"
	"net"
	"testing"
	"time"
)

type dummyAddr string

func (d dummyAddr) Network() string { return "test" }
func (d dummyAddr) String() string  { return string(d) }

type captureConn struct {
	writes [][]byte
}

func (c *captureConn) Read(_ []byte) (int, error) { return 0, io.EOF }
func (c *captureConn) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	c.writes = append(c.writes, cp)
	return len(p), nil
}
func (c *captureConn) Close() error                       { return nil }
func (c *captureConn) LocalAddr() net.Addr                { return dummyAddr("local") }
func (c *captureConn) RemoteAddr() net.Addr               { return dummyAddr("remote") }
func (c *captureConn) SetDeadline(_ time.Time) error      { return nil }
func (c *captureConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *captureConn) SetWriteDeadline(_ time.Time) error { return nil }

func newTestSocketClient(conn net.Conn, key string) *SocketClientImpl {
	block := NewBlockCrypt("xor", key)
	c := NewSocketClient(SocketConfig{
		Address:  "test",
		Protocol: "tcp",
		Block:    block,
	}, &StreamMessagerImpl{timeout: time.Second})
	c.update(conn)
	return c
}

func TestNegotiateLegacyMagic(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	client := newTestSocketClient(c1, "legacy-pre-shared")
	server := newTestSocketClient(c2, "legacy-pre-shared")

	srv := NewSocketServer("test")
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Negotiate(server)
	}()

	if err := client.Negotiate(); err != nil {
		t.Fatalf("client negotiate failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("server negotiate failed: %v", err)
	}
	if client.Key() == "legacy-pre-shared" {
		t.Fatalf("client key should be updated after negotiate")
	}
	if server.Key() == "legacy-pre-shared" {
		t.Fatalf("server key should be updated after negotiate")
	}
	if client.Key() != server.Key() {
		t.Fatalf("negotiated keys mismatch: client=%q server=%q", client.Key(), server.Key())
	}
}

func TestNegotiateMagicV1Network(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	oldResolver := ResolveNetworkCrypt
	defer func() { ResolveNetworkCrypt = oldResolver }()

	network := "team-a"
	preShared := "network-pre-shared"
	ResolveNetworkCrypt = func(name string) *BlockCrypt {
		if name == network {
			return NewBlockCrypt("xor", preShared)
		}
		return nil
	}

	client := newTestSocketClient(c1, preShared)
	client.SetPrivate(ClientCryptDecl{
		Network: network,
		Level:   CryptLevelNetwork,
	})
	server := newTestSocketClient(c2, "global-pre-shared")

	srv := NewSocketServer("test")
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Negotiate(server)
	}()

	if err := client.Negotiate(); err != nil {
		t.Fatalf("client negotiate failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("server negotiate failed: %v", err)
	}
	if client.Key() == preShared {
		t.Fatalf("client key should be updated after negotiate")
	}
	if server.Key() == preShared || server.Key() == "global-pre-shared" {
		t.Fatalf("server key should be updated after negotiate")
	}
	if client.Key() != server.Key() {
		t.Fatalf("negotiated keys mismatch: client=%q server=%q", client.Key(), server.Key())
	}
}

func TestNegotiateWithGlobalLevelUsesLegacyMagic(t *testing.T) {
	conn := &captureConn{}
	client := newTestSocketClient(conn, "global-pre-shared")
	client.SetPrivate(ClientCryptDecl{
		Network: "team-a",
		Level:   CryptLevelGlobal,
	})

	_ = client.Negotiate()
	if len(conn.writes) == 0 {
		t.Fatalf("expected at least one write")
	}
	got := [2]byte{conn.writes[0][0], conn.writes[0][1]}
	if got != MAGIC {
		t.Fatalf("magic mismatch: got=%x want=%x", got, MAGIC)
	}
}

func TestNegotiateWithEmptyNetworkUsesMagicV1(t *testing.T) {
	conn := &captureConn{}
	client := newTestSocketClient(conn, "global-pre-shared")
	client.SetPrivate(ClientCryptDecl{
		Network: "",
		Level:   CryptLevelNetwork,
	})

	_ = client.Negotiate()
	if len(conn.writes) == 0 {
		t.Fatalf("expected at least one write")
	}
	got := [2]byte{conn.writes[0][0], conn.writes[0][1]}
	if got != MAGICv1 {
		t.Fatalf("magic mismatch: got=%x want=%x", got, MAGICv1)
	}
}
