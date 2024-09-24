package config

import (
	"flag"
	"os"
	"path"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type ShadowProxy struct {
	Server     string `json:"server,omitempty"`
	Key        string `json:"key,omitempty"`
	Cipher     string `json:"cipher,omitempty"`
	Password   string `json:"password,omitempty"`
	Plugin     string `json:"plugin,omitempty"`
	PluginOpts string `json:"pluginOpts,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
}

type SocksProxy struct {
	Listen string    `json:"listen,omitempty"`
	Auth   *Password `json:"auth,omitempty"`
}

type HttpForward struct {
	Protocol string   `json:"protocol,omitempty"`
	Server   string   `json:"server,omitempty"`
	Insecure bool     `json:"insecure,omitempty"`
	Match    []string `json:"match,omitempty"`
	Secret   string   `json:"secret,omitempty"`
}

type HttpProxy struct {
	ConfDir  string         `json:"-" yaml:"-"`
	Listen   string         `json:"listen,omitempty"`
	Auth     *Password      `json:"auth,omitempty"`
	Cert     *Cert          `json:"cert,omitempty"`
	Password string         `json:"password,omitempty"`
	Forward  *HttpForward   `json:"forward,omitempty"`
	Backends []*HttpForward `json:"backends,omitempty"`
}

func (h *HttpProxy) Correct() {
	if h.Cert != nil {
		h.Cert.Correct()
	}
	if h.Password != "" && !strings.Contains(h.Password, "/") {
		h.Password = path.Join(h.ConfDir, h.Password)
	}
}

func (h *HttpProxy) FindMatch(domain string, to *HttpForward) int {
	for i, rule := range to.Match {
		if rule == domain {
			return i
		}
	}
	return -1
}

func (h *HttpProxy) FindBackend(remote string) *HttpForward {
	if remote == "" || remote == "null" {
		return h.Forward
	}
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

type TcpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Target []string `json:"target,omitempty"`
}

type Proxy struct {
	Conf    string         `json:"-" yaml:"-"`
	ConfDir string         `json:"-" yaml:"-"`
	Log     Log            `json:"log"`
	Socks   []*SocksProxy  `json:"socks,omitempty" yaml:"socks,omitempty"`
	Http    []*HttpProxy   `json:"http,omitempty" yaml:"http,omitempty"`
	Tcp     []*TcpProxy    `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	Shadow  []*ShadowProxy `json:"shadow,omitempty" yaml:"shadow,omitempty"`
	PProf   string         `json:"pprof,omitempty" yaml:"pprof,omitempty"`
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
	if err := p.Load(); err != nil {
		libol.Error("Proxy.Initialize %s", err)
	}
	p.Correct()
	libol.Debug("Proxy.Initialize %v", p)
}

func (p *Proxy) Correct() {
	p.ConfDir = path.Dir(p.Conf)
	for _, h := range p.Http {
		h.ConfDir = p.ConfDir
		h.Correct()
	}
	p.Log.Correct()
}

func (p *Proxy) Load() error {
	return libol.UnmarshalLoad(p, p.Conf)
}

func (h *Proxy) Save() {
	if err := libol.MarshalSave(&h, h.Conf, true); err != nil {
		libol.Error("Proxy.Save %s %s", h.Conf, err)
	}
}
