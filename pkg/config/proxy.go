package config

import (
	"flag"
	"os"
	"path"

	"github.com/luscis/openlan/pkg/libol"
)

type SocksProxy struct {
	Conf     string     `json:"-" yaml:"-"`
	ConfDir  string     `json:"-" yaml:"-"`
	Listen   string     `json:"listen,omitempty" yaml:"listen,omitempty"`
	Secret   string     `json:"secret,omitempty" yaml:"secret,omitempty"`
	Backends ToForwards `json:"backends,omitempty" yaml:"backends,omitempty"`
	Cert     *Cert      `json:"cert,omitempty" yaml:"cert,omitempty"`
}

func (s *SocksProxy) Initialize() error {
	libol.Info("SocksProxy.Initialize %s", s.Conf)
	if err := s.Load(); err != nil {
		libol.Error("SocksProxy.Initialize %s", err)
		return err
	}
	return nil
}

func (s *SocksProxy) Load() error {
	if s.Conf == "" || libol.FileExist(s.Conf) != nil {
		return libol.NewErr("invalid configure file")
	}
	return libol.UnmarshalLoad(s, s.Conf)
}

func (s *SocksProxy) Correct() {
	if s.Cert != nil {
		s.Cert.Correct()
	}
}

type HttpSocks struct {
	Listen string `json:"listen,omitempty"`
	Cert   *Cert  `json:"-" yaml:"-"`
}

type HttpProxy struct {
	Conf       string      `json:"-" yaml:"-"`
	ConfDir    string      `json:"-" yaml:"-"`
	Listen     string      `json:"listen,omitempty"`
	Secret     string      `json:"secret,omitempty" yaml:"secret,omitempty"`
	Cert       *Cert       `json:"cert,omitempty" yaml:"cert,omitempty"`
	Password   string      `json:"password,omitempty" yaml:"password,omitempty"`
	CaCert     string      `json:"cacert,omitempty" yaml:"cacert,omitempty"`
	Backends   ToForwards  `json:"backends,omitempty" yaml:"backends,omitempty"`
	Socks      *HttpSocks  `json:"socks,omitempty" yaml:"socks,omitempty"`
	SocksProxy *SocksProxy `json:"-" yaml:"-"`
}

func (h *HttpProxy) Initialize() error {
	if h.ConfDir == "" {
		h.ConfDir = path.Dir(os.Args[0])
	}
	libol.Info("HttpProxy.Initialize %s", h.Conf)
	if err := h.Load(); err != nil {
		libol.Error("HttpProxy.Initialize %s", err)
		return err
	}
	h.Correct()
	return nil
}

func (h *HttpProxy) Load() error {
	if h.Conf == "" || libol.FileExist(h.Conf) != nil {
		return libol.NewErr("invalid configure file")
	}
	return libol.UnmarshalLoad(h, h.Conf)
}

func (h *HttpProxy) Correct() {
	if h.Cert != nil {
		h.Cert.Correct()
	}
	if h.Password == "" {
		h.Password = h.Listen + ".pass"
	}
	h.Password = path.Join(h.ConfDir, h.Password)
	if h.CaCert == "" {
		h.CaCert = "ca.crt"
	}
	h.CaCert = path.Join(h.ConfDir, h.CaCert)
	if h.Socks != nil {
		h.SocksProxy = &SocksProxy{
			Listen: h.Socks.Listen,
			Secret: h.Secret,
			Cert:   h.Cert,
		}
	}
	for _, via := range h.Backends {
		via.CaCert = h.CaCert
	}
}

func (h *HttpProxy) FindMatch(domain string, to *ForwardTo) int {
	for i, rule := range to.Match {
		if rule == domain {
			return i
		}
	}
	return -1
}

func (h *HttpProxy) FindBackend(remote string) *ForwardTo {
	for _, to := range h.Backends {
		if to.Server == remote {
			return to
		}
	}
	return nil
}

func (h *HttpProxy) AddMatch(domain, remote string) int {
	to := h.FindBackend(remote)
	if to == nil {
		return -1
	}
	index := h.FindMatch(domain, to)
	if index == -1 {
		to.Match = append(to.Match, domain)
	}
	return 0
}

func (h *HttpProxy) DelMatch(domain, remote string) int {
	to := h.FindBackend(remote)
	if to == nil {
		return -1
	}
	index := h.FindMatch(domain, to)
	if index > -1 {
		to.Match = append(to.Match[:index], to.Match[index+1:]...)
	}
	return index
}

func (h *HttpProxy) Save() {
	if h.Conf == "" {
		return
	}
	if err := libol.MarshalSave(&h, h.Conf, true); err != nil {
		libol.Error("Proxy.Save %s %s", h.Conf, err)
	}
}

type TcpProxy struct {
	Conf   string   `json:"-" yaml:"-"`
	Listen string   `json:"listen,omitempty"`
	Target []string `json:"target,omitempty"`
}

func (t *TcpProxy) Initialize() error {
	libol.Info("TcpProxy.Initialize %s", t.Conf)
	if err := t.Load(); err != nil {
		libol.Error("TcpProxy.Initialize %s", err)
		return err
	}
	return nil
}

func (t *TcpProxy) Load() error {
	if t.Conf == "" || libol.FileExist(t.Conf) != nil {
		return libol.NewErr("invalid configure file")
	}
	return libol.UnmarshalLoad(t, t.Conf)
}

type Proxy struct {
	Conf    string        `json:"-" yaml:"-"`
	ConfDir string        `json:"-" yaml:"-"`
	Log     Log           `json:"log"`
	Socks   []*SocksProxy `json:"socks,omitempty" yaml:"socks,omitempty"`
	Http    []*HttpProxy  `json:"http,omitempty" yaml:"http,omitempty"`
	Tcp     []*TcpProxy   `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	PProf   string        `json:"pprof,omitempty" yaml:"pprof,omitempty"`
}

func NewProxy() *Proxy {
	p := &Proxy{}
	p.Parse()
	p.Initialize()
	return p
}

func (p *Proxy) Parse() {
	flag.StringVar(&p.Log.File, "log:file", "", "Configure log file")
	flag.StringVar(&p.Conf, "conf", "", "The configure file")
	flag.StringVar(&p.PProf, "prof", "", "Http listen for CPU prof")
	flag.IntVar(&p.Log.Verbose, "log:level", 20, "Configure log level")
	flag.Parse()
}

func (p *Proxy) Initialize() {
	if p.Conf == "" {
		p.Conf = path.Dir(os.Args[0]) + "/" + "proxy.json"
	}
	if p.ConfDir == "" {
		p.ConfDir = path.Dir(p.Conf)
	}
	if err := p.Load(); err != nil {
		libol.Error("Proxy.Initialize %s", err)
	}
	p.Correct()
	libol.Debug("Proxy.Initialize %v", p)
}

func (p *Proxy) Correct() {
	p.Log.Correct()
	for _, h := range p.Http {
		h.ConfDir = p.ConfDir
		h.Correct()
	}
	for _, s := range p.Socks {
		s.ConfDir = s.ConfDir
		s.Correct()
	}
}

func (p *Proxy) Load() error {
	return libol.UnmarshalLoad(p, p.Conf)
}

func (p *Proxy) Save() {
	if err := libol.MarshalSave(&p, p.Conf, true); err != nil {
		libol.Error("Proxy.Save %s %s", p.Conf, err)
	}
}
