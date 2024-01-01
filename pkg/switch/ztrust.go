package cswitch

import (
	"fmt"
	"time"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
	cn "github.com/luscis/openlan/pkg/network"
)

type KnockRule struct {
	createAt    int64
	age         int
	protocol    string
	destination string
	port        string
	rule        *cn.IpRule
}

func (r *KnockRule) Id() string {
	return fmt.Sprintf("%s:%s:%s", r.protocol, r.destination, r.port)
}

type ZGuest struct {
	network  string
	username string
	sources  map[string]string
	rules    map[string]*KnockRule
	chain    *cn.FireWallChain
	out      *libol.SubLogger
}

func NewZGuest(network, name string) *ZGuest {
	return &ZGuest{
		network:  network,
		username: name,
		sources:  make(map[string]string, 2),
		rules:    make(map[string]*KnockRule, 1024),
		out:      libol.NewSubLogger(name + "@" + network),
	}
}

func (g *ZGuest) Chain() string {
	return "ZTT_" + g.network + "-" + g.username
}

func (g *ZGuest) Initialize() {
	g.chain = cn.NewFireWallChain(g.Chain(), network.TMangle, "")
}

func (g *ZGuest) Start() {
	g.chain.Install()
}

func (g *ZGuest) AddSource(source string) {
	g.sources[source] = source
}

func (g *ZGuest) DelSource(source string) {
	if _, ok := g.sources[source]; ok {
		delete(g.sources, source)
	}
}

func (g *ZGuest) AddRuleX(rule cn.IpRule) {
	if err := g.chain.AddRuleX(rule); err != nil {
		g.out.Warn("ZTrust.AddRuleX: %s", err)
	}
}

func (g *ZGuest) DelRuleX(rule cn.IpRule) {
	if err := g.chain.DelRuleX(rule); err != nil {
		g.out.Warn("ZTrust.DelRuleX: %s", err)
	}
}

func (g *ZGuest) AddRule(rule *KnockRule) {
	g.rules[rule.Id()] = rule
	g.AddRuleX(cn.IpRule{
		Dest:    rule.destination,
		DstPort: rule.port,
		Proto:   rule.protocol,
		Jump:    "ACCEPT",
		Comment: "Knock at " + time.Now().UTC().String(),
	})
}

func (g *ZGuest) DelRule(rule *KnockRule) {
	if _, ok := g.rules[rule.Id()]; ok {
		delete(g.rules, rule.Id())
	}
	g.DelRuleX(cn.IpRule{
		Proto:   rule.protocol,
		Dest:    rule.destination,
		DstPort: rule.port,
		Jump:    "ACCEPT",
		Comment: "Knock at " + time.Now().Local().String(),
	})
}

func (g *ZGuest) Stop() {
	g.chain.Cancel()
}

type ZTrust struct {
	network string
	expire  int
	guests  map[string]*ZGuest
	chain   *cn.FireWallChain
	out     *libol.SubLogger
}

func NewZTrust(network string, expire int) *ZTrust {
	return &ZTrust{
		network: network,
		expire:  expire,
		out:     libol.NewSubLogger(network),
		guests:  make(map[string]*ZGuest, 32),
	}
}

func (z *ZTrust) Chain() string {
	return "ZTT_" + z.network
}

func (z *ZTrust) Initialize() {
	z.chain = cn.NewFireWallChain(z.Chain(), network.TMangle, "")
	z.chain.AddRule(cn.IpRule{
		Comment: "ZTrust Default",
		Jump:    "DROP",
	})
}

func (z *ZTrust) Knock(name string, protocol, dest, port string, age int) error {
	guest, ok := z.guests[name]
	if !ok {
		return libol.NewErr("Knock: not found %s", name)
	}
	guest.AddRule(&KnockRule{
		protocol:    protocol,
		destination: dest,
		port:        port,
		createAt:    time.Now().Unix(),
		age:         age,
	})
	return nil
}

func (z *ZTrust) Update() {
	//TODO expire knock rules.
}

func (z *ZTrust) AddRuleX(rule cn.IpRule) {
	if err := z.chain.AddRuleX(rule); err != nil {
		z.out.Warn("ZTrust.AddRuleX: %s", err)
	}
}

func (z *ZTrust) DelRuleX(rule cn.IpRule) {
	if err := z.chain.DelRuleX(rule); err != nil {
		z.out.Warn("ZTrust.DelRuleX: %s", err)
	}
}

func (z *ZTrust) AddGuest(name, source string) error {
	z.out.Info("ZTrust.AddGuest: %s %s", name, source)
	if source == "" {
		return libol.NewErr("AddGuest: invalid source")
	}
	guest, ok := z.guests[name]
	if !ok {
		guest = NewZGuest(z.network, name)
		guest.Initialize()
		guest.Start()
		z.guests[name] = guest
	}
	guest.AddSource(source)
	z.AddRuleX(cn.IpRule{
		Source:  source,
		Comment: "User " + guest.username + "@" + guest.network,
		Jump:    guest.Chain(),
		Order:   "-I",
	})
	return nil
}

func (z *ZTrust) DelGuest(name, source string) error {
	if source == "" {
		return libol.NewErr("DelGuest: invalid source")
	}
	guest, ok := z.guests[name]
	if !ok {
		return libol.NewErr("DelGuest: not found %s", name)
	}
	z.out.Info("ZTrust.DelGuest: %s %s", name, source)
	z.DelRuleX(cn.IpRule{
		Source:  source,
		Comment: guest.username + "." + guest.network,
		Jump:    guest.Chain(),
	})
	guest.DelSource(source)
	return nil
}

func (z *ZTrust) Start() {
	z.out.Info("ZTrust.Start")
	z.chain.Install()
}

func (z *ZTrust) Stop() {
	z.out.Info("ZTrust.Stop")
	z.chain.Cancel()
	for _, guest := range z.guests {
		guest.Stop()
	}
}
