package api

import (
	"net"

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

type ACLer interface {
	AddRule(rule *schema.ACLRule) error
	DelRule(rule *schema.ACLRule) error
	ListRules(call func(obj schema.ACLRule))
	Save()
}

type ZTruster interface {
	AddGuest(name, source string) error
	DelGuest(name, source string) error
	Knock(name string, protocol, dest, port string, age int) error
	ListGuest(call func(obj schema.ZGuest))
	ListKnock(name string, call func(obj schema.KnockRule))
}

type Router interface {
	AddRoute(route *schema.PrefixRoute, switcher Switcher) error
	DelRoute(route *schema.PrefixRoute, switcher Switcher) error
	SaveRoute()
}

type Vpner interface {
	RestartVpn()
}

type Qoser interface {
	AddQosUser(name string, inSpeed float64) error
	UpdateQosUser(name string, inSpeed float64) error
	DelQosUser(name string) error
	ListQosUsers(call func(obj schema.Qos))
	Save()
}

type Outputer interface {
	AddOutput(data schema.Output)
	DelOutput(device string)
	SaveOutput()
}

type Networker interface {
	String() string
	ID() string
	Initialize()
	Start(v Switcher)
	Stop()
	Bridge() cn.Bridger
	Config() *co.Network
	Subnet() *net.IPNet
	Reload(v Switcher)
	Provider() string
	ZTruster() ZTruster
	Qoser() Qoser
	IfAddr() string
	ACLer() ACLer
	Outputer
	Router
	Vpner
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
