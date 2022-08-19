package config

import (
	"flag"
	"github.com/luscis/openlan/pkg/libol"
	"runtime"
	"strings"
)

type Interface struct {
	Name     string `json:"name,omitempty"`
	IPMtu    int    `json:"mtu"`
	Address  string `json:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty"`
	Provider string `json:"provider,omitempty"`
	Cost     int    `json:"cost,omitempty"`
}

type Point struct {
	File        string    `json:"file,omitempty"`
	Alias       string    `json:"alias,omitempty"`
	Connection  string    `json:"connection"`
	Timeout     int       `json:"timeout"`
	Username    string    `json:"username,omitempty"`
	Network     string    `json:"network"`
	Password    string    `json:"password,omitempty"`
	Protocol    string    `json:"protocol,omitempty"`
	Interface   Interface `json:"interface"`
	Log         Log       `json:"log"`
	Http        *Http     `json:"http,omitempty"`
	Crypt       *Crypt    `json:"crypt,omitempty"`
	PProf       string    `json:"pprof,omitempty"`
	RequestAddr bool      `json:"requestAddr,omitempty"`
	ByPass      bool      `json:"bypass,omitempty"`
	SaveFile    string    `json:"-"`
	Queue       *Queue    `json:"queue,omitempty"`
	Terminal    string    `json:"-"`
	Cert        *Cert     `json:"cert,omitempty"`
	StatusFile  string    `json:"status,omitempty"`
	PidFile     string    `json:"pid,omitempty"`
}

func DefaultPoint() *Point {
	obj := &Point{
		Alias:      "",
		Connection: "xx.openlan.net",
		Network:    "default",
		Protocol:   "tcp", // udp, kcp, tcp, tls, ws and wss etc.
		Timeout:    60,
		Log: Log{
			File:    "./point.log",
			Verbose: libol.INFO,
		},
		Interface: Interface{
			IPMtu:    1500,
			Provider: "kernel",
			Name:     "",
			Cost:     1000,
		},
		SaveFile:    "./point.json",
		RequestAddr: true,
		Crypt:       &Crypt{},
		Cert:        &Cert{},
		Terminal:    "on",
	}
	obj.Correct(nil)
	return obj
}

func NewPoint() *Point {
	obj := DefaultPoint()
	p := &Point{
		RequestAddr: true,
		Crypt:       obj.Crypt,
		Cert:        obj.Cert,
		Interface:   obj.Interface,
	}
	p.Flags()
	p.Parse()
	p.Initialize()
	return p
}

func (ap *Point) Flags() {
	obj := DefaultPoint()
	flag.StringVar(&ap.Alias, "alias", obj.Alias, "Alias for this point")
	flag.StringVar(&ap.Terminal, "terminal", obj.Terminal, "Run interactive terminal")
	flag.StringVar(&ap.Connection, "conn", obj.Connection, "Connection access to")
	flag.StringVar(&ap.Log.File, "log:file", obj.Log.File, "File log saved to")
	flag.IntVar(&ap.Log.Verbose, "log:level", obj.Log.Verbose, "Log level value")
	flag.StringVar(&ap.SaveFile, "conf", obj.SaveFile, "The configuration file")
	flag.StringVar(&ap.PProf, "pprof", obj.PProf, "Http listen for pprof debug")
	flag.StringVar(&ap.Cert.CaFile, "cacert", obj.Cert.CaFile, "CA certificate file")
	flag.StringVar(&ap.PidFile, "pid", obj.PidFile, "Write pid to file")
}

func (ap *Point) Parse() {
	flag.Parse()
}

func (ap *Point) Id() string {
	return ap.Connection + ":" + ap.Network
}

func (ap *Point) Initialize() {
	if err := ap.Load(); err != nil {
		libol.Warn("NewPoint.Initialize %s", err)
	}
	ap.Default()
	libol.SetLogger(ap.Log.File, ap.Log.Verbose)
}

func (ap *Point) Correct(obj *Point) {
	if ap.Alias == "" {
		ap.Alias = GetAlias()
	}
	if ap.Network == "" {
		if strings.Contains(ap.Username, "@") {
			ap.Network = strings.SplitN(ap.Username, "@", 2)[1]
		} else if obj != nil {
			ap.Network = obj.Network
		}
	}
	CorrectAddr(&ap.Connection, 10002)
	if runtime.GOOS == "darwin" {
		ap.Interface.Provider = "tun"
	}
	if ap.Protocol == "tls" || ap.Protocol == "wss" {
		if ap.Cert == nil && obj != nil {
			ap.Cert = obj.Cert
		}
	}
	if ap.Protocol == "" {
		ap.Protocol = obj.Protocol
	}
	if ap.Cert != nil {
		if ap.Cert.Dir == "" {
			ap.Cert.Dir = "."
		}
		ap.Cert.Correct()
	}
	if ap.Timeout == 0 {
		ap.Timeout = obj.Timeout
	}
	if ap.Interface.Cost == 0 {
		ap.Interface.Cost = obj.Interface.Cost
	}
	if ap.Interface.IPMtu == 0 {
		ap.Interface.IPMtu = obj.Interface.IPMtu
	}
}

func (ap *Point) Default() {
	obj := DefaultPoint()
	ap.Correct(obj)
	if ap.Queue == nil {
		ap.Queue = &Queue{}
	}
	ap.Queue.Default()
	//reset zero value to default
	if ap.Crypt != nil {
		ap.Crypt.Default()
	}
}

func (ap *Point) Load() error {
	if err := libol.FileExist(ap.SaveFile); err == nil {
		return libol.UnmarshalLoad(ap, ap.SaveFile)
	}
	return nil
}
