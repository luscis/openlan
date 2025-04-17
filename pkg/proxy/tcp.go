package proxy

import (
	"io"
	"net"
	"time"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

type TcpProxy struct {
	listen   string
	target   []string
	listener net.Listener
	out      *libol.SubLogger
	rr       uint64
}

func NewTcpProxy(cfg *config.TcpProxy) *TcpProxy {
	return &TcpProxy{
		listen: cfg.Listen,
		target: cfg.Target,
		out:    libol.NewSubLogger(cfg.Listen),
	}
}

func (t *TcpProxy) Initialize() {
}

func (t *TcpProxy) tunnel(src net.Conn, dst net.Conn) {
	defer dst.Close()
	defer src.Close()
	t.out.Info("TcpProxy.tunnel %s -> %s", src.RemoteAddr(), dst.RemoteAddr())
	wait := libol.NewWaitOne(2)
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(dst, src); err != nil {
			t.out.Debug("TcpProxy.tunnel from ws %s", err)
		}
	})
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(src, dst); err != nil {
			t.out.Debug("TcpProxy.tunnel from target %s", err)
		}
	})
	wait.Wait()
	t.out.Debug("TcpProxy.tunnel %s exit", dst.RemoteAddr())
}

func (t *TcpProxy) loadBalance(fail int) string {
	size := len(t.target)
	if fail < size {
		i := t.rr % uint64(size)
		t.rr++
		return t.target[i]
	}
	return ""
}

func (t *TcpProxy) Start() {
	var listen net.Listener
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Do(func() error {
		var err error
		listen, err = net.Listen("tcp", t.listen)
		if err != nil {
			t.out.Warn("TcpProxy.Start %s", err)
		}
		return err
	})
	t.listener = listen
	t.out.Info("TcpProxy.Start: %s", t.target)
	libol.Go(func() {
		defer listen.Close()
		for {
			conn, err := listen.Accept()
			if err != nil {
				t.out.Error("TcpServer.Accept: %s", err)
				break
			}
			// connect target and pipe it.
			fail := 0
			for {
				backend := t.loadBalance(fail)
				if backend == "" {
					break
				}
				target, err := net.Dial("tcp", backend)
				if err != nil {
					t.out.Error("TcpProxy.Accept %s", err)
					fail++
					continue
				}
				libol.Go(func() {
					t.tunnel(conn, target)
				})
				break
			}
		}
	})
	return
}

func (t *TcpProxy) Stop() {
	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}
	t.out.Info("TcpProxy.Stop")
}

func (t *TcpProxy) Save() {
}
