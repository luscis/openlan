package cswitch

import (
	"fmt"
	"time"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
	cn "github.com/luscis/openlan/pkg/network"
)

type KnockRule struct {
	createAt int64
	protocol string
	dest     string
	destPort string
	rule     *cn.IpRule
}

func (r *KnockRule) Id() string {
	return fmt.Sprintf("%s:%s:%s", r.protocol, r.dest, r.destPort)
}

type ZeroGuest struct {
	network  string
	username string
	device   string
	sources  map[string]string
	rules    map[string]*KnockRule
	chain    *cn.FireWallChain
}

func NewZeroGuest(network, name string) *ZeroGuest {
	return &ZeroGuest{
		network:  network,
		username: name,
		sources:  make(map[string]string, 2),
		rules:    make(map[string]*KnockRule, 1024),
	}
}

func (z *ZeroGuest) Chain() string {
	return "ZTT_" + z.network + "_" + z.username
}

func (z *ZeroGuest) Initialize() {
	z.chain = cn.NewFireWallChain(z.Chain(), network.TMangle, "")
}

func (z *ZeroGuest) AddSource(source string) {
	z.sources[source] = source
}

func (z *ZeroGuest) DelSource(source string) {
	if _, ok := z.sources[source]; ok {
		delete(z.sources, source)
	}
}

func (z *ZeroGuest) AddRule(rule *KnockRule) {
	z.rules[rule.Id()] = rule
	z.chain.AddRuleX(cn.IpRule{
		Dest:    rule.dest,
		DstPort: rule.destPort,
		Jump:    "ACCEPT",
		Comment: time.Now().Local().String(),
	})
}

func (z *ZeroGuest) DelRule(rule *KnockRule) {
	if _, ok := z.rules[rule.Id()]; ok {
		delete(z.rules, rule.Id())
	}
}

type ZeroTrust struct {
	network string
	expire  int
	guests  map[string]*ZeroGuest
	chain   *cn.FireWallChain
	out     *libol.SubLogger
}

func NewZeroTrust(network string, expire int) *ZeroTrust {
	return &ZeroTrust{
		network: network,
		expire:  expire,
		out:     libol.NewSubLogger(network),
		guests:  make(map[string]*ZeroGuest, 32),
	}
}

func (z *ZeroTrust) Chain() string {
	return "ZTT_" + z.network
}

func (z *ZeroTrust) Initialize() {
	z.chain = cn.NewFireWallChain(z.Chain(), network.TMangle, "")
	z.chain.AddRule(cn.IpRule{
		Comment: "Zero Trust Default",
		Jump:    "DROP",
	})
}

func (z *ZeroTrust) Knock(name string, dest, protocol, destPort string) {
	guest, ok := z.guests[name]
	if !ok {
		z.out.Warn("ZeroTrust.Knock: not found %s", name)
		return
	}
	guest.AddRule(&KnockRule{
		protocol: protocol,
		dest:     dest,
		destPort: destPort,
		createAt: time.Now().Unix(),
	})
}

func (z *ZeroTrust) Update() {
	//TODO expire knock rules.
}

func (z *ZeroTrust) AddGuest(name, device, source string) {
	guest, ok := z.guests[name]
	if !ok {
		guest = NewZeroGuest(z.network, name)
		guest.Initialize()
		z.guests[name] = guest
	}
	guest.AddSource(source)
	z.chain.AddRuleX(cn.IpRule{
		Input:   device,
		Source:  source,
		Comment: guest.username + "." + guest.network,
		Jump:    guest.Chain(),
		Order:   "-I",
	})
}

func (z *ZeroTrust) DelGuest(name, device, source string) {
	guest, ok := z.guests[name]
	if !ok {
		return
	}
	z.chain.DelRuleX(cn.IpRule{
		Input:   device,
		Source:  source,
		Comment: guest.username + "." + guest.network,
		Jump:    guest.Chain(),
	})
	guest.DelSource(source)
}

func (z *ZeroTrust) Start() {
	z.out.Info("ZeroTrust.Start")
	z.chain.Install()
}

func (z *ZeroTrust) Stop() {
	z.out.Info("ZeroTrust.Stop")
	z.chain.Cancel()
}
