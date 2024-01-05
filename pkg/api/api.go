package api

import (
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type Switcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	Config() *co.Switch
	Server() libol.SocketServer
	Reload()
	Save()
}

func NewWorkerSchema(s Switcher) schema.Worker {
	protocol := ""
	if cfg := s.Config(); cfg != nil {
		protocol = cfg.Protocol
	}
	return schema.Worker{
		UUID:     s.UUID(),
		Uptime:   s.UpTime(),
		Alias:    s.Alias(),
		Protocol: protocol,
	}
}

type ZTruster interface {
	AddGuest(name, source string) error
	DelGuest(name, source string) error
	Knock(name string, protocol, dest, port string, age int) error
	ListGuest(call func(obj schema.ZGuest))
	ListKnock(name string, call func(obj schema.KnockRule))
}

type Networker interface {
	String() string
	ID() string
	Initialize()
	Start(v Switcher)
	Stop()
	Bridge() cn.Bridger
	Config() *co.Network
	Subnet() string
	Reload(v Switcher)
	Provider() string
	ZTruster() ZTruster
	IfAddr() string
}

var workers = make(map[string]Networker)

func AddWorker(name string, obj Networker) {
	workers[name] = obj
}

func GetWorker(name string) Networker {
	return workers[name]
}

func ListWorker(call func(w Networker)) {
	for _, worker := range workers {
		call(worker)
	}
}
