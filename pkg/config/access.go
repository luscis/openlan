package config

import (
	"flag"
	"runtime"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type Interface struct {
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	IPMtu    int    `json:"mtu,omitempty" yaml:"mtu,omitempty"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty" yaml:"bridge,omitempty"`
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Cost     int    `json:"cost,omitempty" yaml:"cost,omitempty"`
}

type Access struct {
	File        string    `json:"file,omitempty" yaml:"file,omitempty"`
	Alias       string    `json:"alias,omitempty" yaml:"alias,omitempty"`
	Connection  string    `json:"connection" yaml:"connection"`
	Timeout     int       `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Username    string    `json:"username,omitempty" yaml:"username,omitempty"`
	Network     string    `json:"network,omitempty" yaml:"network,omitempty"`
	Password    string    `json:"password,omitempty" yaml:"password,omitempty"`
	Protocol    string    `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Interface   Interface `json:"interface,omitempty" yaml:"interface,omitempty"`
	Log         Log       `json:"log,omitempty" yaml:"log,omitempty"`
	Http        *Http     `json:"http,omitempty" yaml:"http,omitempty"`
	Crypt       *Crypt    `json:"crypt,omitempty" yaml:"crypt,omitempty"`
	PProf       string    `json:"pprof,omitempty" yaml:"pprof,omitempty"`
	RequestAddr bool      `json:"requestAddr" yaml:"requestAddr"`
	Conf        string    `json:"-" yaml:"-"`
	Queue       *Queue    `json:"queue,omitempty" yaml:"queue,omitempty"`
	Cert        *Cert     `json:"cert,omitempty" yaml:"cert,omitempty"`
	StatusFile  string    `json:"status,omitempty" yaml:"status,omitempty"`
	PidFile     string    `json:"pid,omitempty" yaml:"pid,omitempty"`
	Forward     []string  `json:"forward,omitempty" yaml:"forward,omitempty"`
}

func (i *Interface) Correct() {
	if i.Provider == "" {
		i.Provider = "kernel"
	}
	if i.IPMtu == 0 {
		i.IPMtu = 1500
	}
}

func NewAccess() *Access {
	p := &Access{RequestAddr: true}
	p.Parse()
	p.Initialize()
	return p
}

func (ap *Access) Parse() {
	flag.StringVar(&ap.Alias, "alias", "", "Alias for this Access")
	flag.StringVar(&ap.Log.File, "log:file", "", "File log saved to")
	flag.StringVar(&ap.Conf, "conf", "", "The configuration file")
	flag.Parse()
}

func (ap *Access) Id() string {
	return ap.Connection + ":" + ap.Network
}

func (ap *Access) Initialize() error {
	if err := ap.Load(); err != nil {
		libol.Warn("NewAccess.Initialize %s", err)
		return err
	}
	ap.Correct()
	libol.SetLogger(ap.Log.File, ap.Log.Verbose)
	return nil
}

func (ap *Access) Correct() {
	if ap.Alias == "" {
		ap.Alias = GetAlias()
	}
	if ap.Network == "" {
		if strings.Contains(ap.Username, "@") {
			ap.Network = strings.SplitN(ap.Username, "@", 2)[1]
		}
	}
	SetListen(&ap.Connection, 10002)
	if runtime.GOOS == "darwin" {
		ap.Interface.Provider = "tun"
	}
	if ap.Protocol == "tls" || ap.Protocol == "wss" {
		if ap.Cert == nil {
			ap.Cert = &Cert{}
		}
	}
	if ap.Protocol == "" {
		ap.Protocol = "tcp"
	}
	if ap.Cert != nil {
		ap.Cert.Correct()
	}
	if ap.Crypt == nil {
		ap.Crypt = &Crypt{}
	}
	ap.Crypt.Correct()
	if ap.Timeout == 0 {
		ap.Timeout = 60
	}
	ap.Interface.Correct()
	ap.Log.Correct()
	if ap.Queue == nil {
		ap.Queue = &Queue{}
	}
	ap.Queue.Correct()
}

func (ap *Access) Load() error {
	if err := libol.FileExist(ap.Conf); err == nil {
		return libol.UnmarshalLoad(ap, ap.Conf)
	}
	return nil
}
