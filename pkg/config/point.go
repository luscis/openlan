package config

import (
	"flag"
	"runtime"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type Interface struct {
	Name     string `json:"name,omitempty"`
	IPMtu    int    `json:"mtu,omitempty"`
	Address  string `json:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty"`
	Provider string `json:"provider,omitempty"`
	Cost     int    `json:"cost,omitempty"`
}

type Point struct {
	File        string    `json:"file,omitempty"`
	Alias       string    `json:"alias,omitempty"`
	Connection  string    `json:"connection"`
	Timeout     int       `json:"timeout,omitempty"`
	Username    string    `json:"username,omitempty"`
	Network     string    `json:"network,omitempty"`
	Password    string    `json:"password,omitempty"`
	Protocol    string    `json:"protocol,omitempty"`
	Interface   Interface `json:"interface,omitempty"`
	Log         Log       `json:"log,omitempty"`
	Http        *Http     `json:"http,omitempty"`
	Crypt       *Crypt    `json:"crypt,omitempty"`
	PProf       string    `json:"pprof,omitempty"`
	RequestAddr bool      `json:"requestAddr"`
	Conf        string    `json:"-"`
	Queue       *Queue    `json:"queue,omitempty"`
	Cert        *Cert     `json:"cert,omitempty"`
	StatusFile  string    `json:"status,omitempty" yaml:"status,omitempty" `
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

func NewPoint() *Point {
	p := &Point{RequestAddr: true}
	p.Parse()
	p.Initialize()
	return p
}

func (ap *Point) Parse() {
	flag.StringVar(&ap.Alias, "alias", "", "Alias for this point")
	flag.StringVar(&ap.Log.File, "log:file", "", "File log saved to")
	flag.StringVar(&ap.Conf, "conf", "", "The configuration file")
	flag.Parse()
}

func (ap *Point) Id() string {
	return ap.Connection + ":" + ap.Network
}

func (ap *Point) Initialize() error {
	if err := ap.Load(); err != nil {
		libol.Warn("NewPoint.Initialize %s", err)
		return err
	}
	ap.Correct()
	libol.SetLogger(ap.Log.File, ap.Log.Verbose)
	return nil
}

func (ap *Point) Correct() {
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

func (ap *Point) Load() error {
	if err := libol.FileExist(ap.Conf); err == nil {
		return libol.UnmarshalLoad(ap, ap.Conf)
	}
	return nil
}
