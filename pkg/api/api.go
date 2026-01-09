package api

import (
	"net"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type RateLimitApi interface {
	AddRate(device string, mbit int)
	DelRate(device string)
}

type LdapApi interface {
	AddLdap(value schema.LDAP) error
	DelLdap()
}

type SwitchApi interface {
	UUID() string
	UpTime() int64
	Alias() string
	Config() *co.Switch
	Server() libol.SocketServer
	Reload()
	Save()
	AddNetwork(string)
	DelNetwork(string)
	SaveNetwork(string)
	RateLimitApi
	UpdateCert(schema.VersionCert)
	GetCert() schema.VersionCert
	LdapApi
}

func NewWorkerSchema(s SwitchApi) schema.Worker {
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

type ACLApi interface {
	AddRule(rule *schema.ACLRule) error
	DelRule(rule *schema.ACLRule) error
	ListRules(call func(obj schema.ACLRule))
	SaveRule()
}

type ZTrustApi interface {
	AddGuest(name, source string) error
	DelGuest(name, source string) error
	Knock(name string, protocol, dest, port string, age int) error
	ListGuest(call func(obj schema.ZGuest))
	ListKnock(name string, call func(obj schema.KnockRule))
}

type RouteApi interface {
	AddRoute(route *schema.PrefixRoute, switchApi SwitchApi) error
	DelRoute(route *schema.PrefixRoute, switchApi SwitchApi) error
	ListRoute(call func(obj schema.PrefixRoute))
	SaveRoute()
}

type VPNApi interface {
	AddVPN(schema.OpenVPN) error
	DelVPN()
	StartVPN()
	AddVPNClient(name, local string) error
	DelVPNClient(name string) error
	ListClients(call func(name, local string))
	KillVPNClient(name string) error
}

type QosApi interface {
	AddQos(name string, inSpeed float64) error
	UpdateQos(name string, inSpeed float64) error
	DelQos(name string) error
	ListQos(call func(obj schema.Qos))
	SaveQos()
}

type OutputApi interface {
	AddOutput(data schema.Output)
	DelOutput(data schema.Output)
	SaveOutput()
}

type FindHopApi interface {
	AddHop(data schema.FindHop) error
	DelHop(data schema.FindHop) error
	ListHop(call func(obj schema.FindHop))
	SaveHop()
}

type SNATApi interface {
	SetSnat(string)
}

type DNATApi interface {
	AddDnat(data schema.DNAT) error
	DelDnat(data schema.DNAT) error
	ListDnat(call func(obj schema.DNAT))
}

type SupeApi interface {
	String() string
	ID() string
	Initialize()
	Start(v SwitchApi)
	Stop()
	Reload(v SwitchApi)
}

type NetworkApi interface {
	SupeApi
	Config() *co.Network
	Subnet() *net.IPNet
	Provider() string
	IfAddr() string
	SetMss(mss int)
	OutputApi
	RouteApi
	VPNApi
	Bridger() cn.Bridger
	ZTruster() ZTrustApi
	Qoser() QosApi
	ACLer() ACLApi
	FindHoper() FindHopApi
	EnableZTrust()
	DisableZTrust()
	SNATApi
	DNATApi
	AddAddress(string)
	DelAddress()
}

type IPSecApi interface {
	AddTunnel(data schema.IPSecTunnel)
	DelTunnel(data schema.IPSecTunnel)
	StartTunnel(data schema.IPSecTunnel)
	ListTunnels(call func(obj schema.IPSecTunnel))
}

type BgpApi interface {
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

type CeciApi interface {
	AddTcp(data schema.CeciTcp)
	DelTcp(data schema.CeciTcp)
}

type RouterApi interface {
	AddTunnel(data schema.RouterTunnel) error
	DelTunnel(data schema.RouterTunnel) error
	AddPrivate(data string) error
	DelPrivate(data string) error
	AddInterface(data schema.RouterInterface) error
	DelInterface(data schema.RouterInterface) error
}

type callApi struct {
	ipsecApi  IPSecApi
	bgpApi    BgpApi
	ceciApi   CeciApi
	routerApi RouterApi
	workers   map[string]NetworkApi
}

func (i *callApi) AddWorker(name string, obj NetworkApi) {
	i.workers[name] = obj
}

func (i *callApi) GetWorker(name string) NetworkApi {
	return i.workers[name]
}

func (i *callApi) ListWorker(call func(w NetworkApi)) {
	for _, w := range i.workers {
		call(w)
	}
}

func (i *callApi) SetIPSecApi(value IPSecApi) {
	i.ipsecApi = value
}

func (i *callApi) SetBgpApi(value BgpApi) {
	i.bgpApi = value
}

func (i *callApi) SetCeciApi(value CeciApi) {
	i.ceciApi = value
}

func (i *callApi) SetRouterApi(value RouterApi) {
	i.routerApi = value
}

var Call = &callApi{
	workers: make(map[string]NetworkApi),
}
