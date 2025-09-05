package api

import (
	"net"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type Rater interface {
	AddRate(device string, mbit int)
	DelRate(device string)
}

type Switcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	Config() *co.Switch
	Server() libol.SocketServer
	Reload()
	Save()
	AddNetwork(network string)
	DelNetwork(network string)
	SaveNetwork(network string)
	Rater
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
	SaveRule()
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
	ListRoute(call func(obj schema.PrefixRoute))
	SaveRoute()
}

type VPNer interface {
	RestartVPN()
}

type Qoser interface {
	AddQos(name string, inSpeed float64) error
	UpdateQos(name string, inSpeed float64) error
	DelQos(name string) error
	ListQos(call func(obj schema.Qos))
	SaveQos()
}

type Outputer interface {
	AddOutput(data schema.Output)
	DelOutput(data schema.Output)
	SaveOutput()
}

type FindHoper interface {
	AddHop(data schema.FindHop) error
	DelHop(data schema.FindHop) error
	ListHop(call func(obj schema.FindHop))
	SaveHop()
}

type Super interface {
	String() string
	ID() string
	Initialize()
	Start(v Switcher)
	Stop()
	Reload(v Switcher)
}

type Networker interface {
	Super
	Config() *co.Network
	Subnet() *net.IPNet
	Provider() string
	IfAddr() string
	SetMss(mss int)
	Outputer
	Router
	VPNer
	Bridger() cn.Bridger
	ZTruster() ZTruster
	Qoser() Qoser
	ACLer() ACLer
	FindHoper() FindHoper
	EnableZTrust()
	DisableZTrust()
	EnableSnat()
	DisableSnat()
}

type IPSecer interface {
	AddTunnel(data schema.IPSecTunnel)
	DelTunnel(data schema.IPSecTunnel)
	RestartTunnel(data schema.IPSecTunnel)
	ListTunnels(call func(obj schema.IPSecTunnel))
}

type Bgper interface {
	Enable(data schema.Bgp)
	Disable()
	Get() *schema.Bgp
	AddNeighbor(data schema.BgpNeighbor)
	DelNeighbor(data schema.BgpNeighbor)
	AddReceives(data schema.BgpPrefix)
	DelReceives(data schema.BgpPrefix)
	AddAdvertis(data schema.BgpPrefix)
	DelAdvertis(data schema.BgpPrefix)
}

type APICall struct {
	secer   IPSecer
	bgper   Bgper
	workers map[string]Networker
}

func (i *APICall) AddWorker(name string, obj Networker) {
	i.workers[name] = obj
}

func (i *APICall) GetWorker(name string) Networker {
	return i.workers[name]
}

func (i *APICall) ListWorker(call func(w Networker)) {
	for _, worker := range i.workers {
		call(worker)
	}
}

func (i *APICall) SetIPSecer(value IPSecer) {
	i.secer = value
}

func (i *APICall) GetIPSecer() IPSecer {
	return i.secer
}

func (i *APICall) SetBgper(value Bgper) {
	i.bgper = value
}

func (i *APICall) GetBgper() Bgper {
	return i.bgper
}

var Call = &APICall{
	workers: make(map[string]Networker),
}
