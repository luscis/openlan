package proxy

import (
	"os"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy/ss"
)

type Proxyer interface {
	Initialize()
	Start()
	Stop()
	Save()
}

type Proxy struct {
	cfg    *config.Proxy
	tcp    map[string]*TcpProxy
	socks  map[string]*SocksProxy
	http   map[string]*HttpProxy
	shadow map[string]*ss.ShadowSocks
}

func NewProxy(cfg *config.Proxy) *Proxy {
	return &Proxy{
		cfg:    cfg,
		socks:  make(map[string]*SocksProxy, 32),
		tcp:    make(map[string]*TcpProxy, 32),
		http:   make(map[string]*HttpProxy, 32),
		shadow: make(map[string]*ss.ShadowSocks, 32),
	}
}

func (p *Proxy) Initialize() {
	if p.cfg == nil {
		return
	}
	for _, c := range p.cfg.Socks {
		s := NewSocksProxy(c)
		if s == nil {
			continue
		}
		p.socks[c.Listen] = s
	}
	for _, c := range p.cfg.Tcp {
		p.tcp[c.Listen] = NewTcpProxy(c)
	}
	for _, c := range p.cfg.Http {
		if c == nil || c.Listen == "" {
			continue
		}
		h := NewHttpProxy(c, p)
		p.http[c.Listen] = h
	}
	for _, c := range p.cfg.Shadow {
		if c == nil || c.Server == "" {
			continue
		}
		h := ss.NewShadowSocks(c)
		p.shadow[c.Server] = h
	}
}

func (p *Proxy) Start() {
	if p.cfg == nil {
		return
	}
	libol.Info("Proxy.Start")
	for _, s := range p.socks {
		s.Start()
	}
	for _, t := range p.tcp {
		t.Start()
	}
	for _, h := range p.http {
		h.Start()
	}
	for _, s := range p.shadow {
		s.Start()
	}
}

func (p *Proxy) Stop() {
	if p.cfg == nil {
		return
	}
	libol.Info("Proxy.Stop")
	for _, t := range p.tcp {
		t.Stop()
	}
	for _, s := range p.shadow {
		s.Stop()
	}
}

func (p *Proxy) Save() {
	p.cfg.Save()
}

func init() {
	// HTTP/2.0 not support upgrade for Hijacker
	if err := os.Setenv("GODEBUG", "http2server=0"); err != nil {
		libol.Warn("proxy.init %s")
	}
}
