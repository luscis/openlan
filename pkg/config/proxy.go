package config

import (
	"flag"
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
	Listen string   `json:"listen,omitempty"`
	Auth   Password `json:"auth,omitempty"`
}

type HttpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Auth   Password `json:"auth,omitempty"`
	Cert   *Cert    `json:"cert,omitempty"`
}

type TcpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Target []string `json:"target,omitempty"`
}

type Proxy struct {
	Conf   string         `json:"-"`
	Log    Log            `json:"log"`
	Socks  []*SocksProxy  `json:"socks,omitempty"`
	Http   []*HttpProxy   `json:"http,omitempty"`
	Tcp    []*TcpProxy    `json:"tcp,omitempty"`
	Shadow []*ShadowProxy `json:"shadow,omitempty"`
	PProf  string         `json:"pprof"`
}

func DefaultProxy() *Proxy {
	obj := &Proxy{
		Log: Log{
			File:    LogFile("openlan-proxy.log"),
			Verbose: libol.INFO,
		},
	}
	obj.Correct(nil)
	return obj
}

func NewProxy() *Proxy {
	p := &Proxy{}
	p.Flags()
	p.Parse()
	p.Initialize()
	if Manager.Proxy == nil {
		Manager.Proxy = p
	}
	return p
}

func (p *Proxy) Flags() {
	obj := DefaultProxy()
	flag.StringVar(&p.Log.File, "log:file", obj.Log.File, "Configure log file")
	flag.StringVar(&p.Conf, "conf", obj.Conf, "The configure file")
	flag.StringVar(&p.PProf, "prof", obj.PProf, "Http listen for CPU prof")
	flag.IntVar(&p.Log.Verbose, "log:level", obj.Log.Verbose, "Configure log level")
}

func (p *Proxy) Parse() {
	flag.Parse()
}

func (p *Proxy) Initialize() {
	if err := p.Load(); err != nil {
		libol.Error("Proxy.Initialize %s", err)
	}
	p.Default()
	libol.Debug("Proxy.Initialize %v", p)
}

func (p *Proxy) Correct(obj *Proxy) {
	for _, h := range p.Http {
		if h.Cert != nil {
			h.Cert.Correct()
		}
	}
}

func (p *Proxy) Default() {
	p.Correct(nil)
}

func (p *Proxy) Load() error {
	return libol.UnmarshalLoad(p, p.Conf)
}
