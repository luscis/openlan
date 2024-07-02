package config

import (
	"flag"
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
	ConfDir  string       `json:"-"`
	Listen   string       `json:"listen,omitempty"`
	Auth     *Password    `json:"auth,omitempty"`
	Cert     *Cert        `json:"cert,omitempty"`
	Password string       `json:"password,omitempty"`
	Forward  *HttpForward `json:"forward,omitempty"`
}

func (h *HttpProxy) Correct() {
	if h.Cert != nil {
		h.Cert.Correct()
	}
	if h.Password != "" && !strings.Contains(h.Password, "/") {
		h.Password = path.Join(h.ConfDir, h.Password)
	}
}

type TcpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Target []string `json:"target,omitempty"`
}

type Proxy struct {
	Conf    string         `json:"-"`
	ConfDir string         `json:"-"`
	Log     Log            `json:"log"`
	Socks   []*SocksProxy  `json:"socks,omitempty"`
	Http    []*HttpProxy   `json:"http,omitempty"`
	Tcp     []*TcpProxy    `json:"tcp,omitempty"`
	Shadow  []*ShadowProxy `json:"shadow,omitempty"`
	PProf   string         `json:"pprof"`
}

func NewProxy() *Proxy {
	p := &Proxy{}
	p.Parse()
	p.Initialize()
	return p
}

func (p *Proxy) Parse() {
	flag.StringVar(&p.Log.File, "log:file", "", "Configure log file")
	flag.StringVar(&p.Conf, "conf", "proxy.json", "The configure file")
	flag.StringVar(&p.PProf, "prof", "", "Http listen for CPU prof")
	flag.IntVar(&p.Log.Verbose, "log:level", 20, "Configure log level")
	flag.Parse()
}

func (p *Proxy) Initialize() {
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
