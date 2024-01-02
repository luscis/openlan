package cswitch

import (
	"fmt"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type KnockRule struct {
	createAt    time.Time
	age         int64
	protocol    string
	destination string
	port        string
	rule        *cn.IpRule
}

func (r *KnockRule) Id() string {
	return fmt.Sprintf("%s:%s:%s", r.protocol, r.destination, r.port)
}

func (r *KnockRule) Expire() bool {
	now := time.Now()
	if r.createAt.Unix()+int64(r.age) < now.Unix() {
		return true
	}
	return false
}

func (r *KnockRule) Rule() cn.IpRule {
	if r.rule == nil {
		r.rule = &cn.IpRule{
			Dest:    r.destination,
			DstPort: r.port,
			Proto:   r.protocol,
			Jump:    "ACCEPT",
			Comment: "Knock at " + r.createAt.UTC().String(),
		}
	}
	return *r.rule
}

type ZGuest struct {
	network  string
	username string
	sources  map[string]string
	rules    map[string]*KnockRule
	chain    *cn.FireWallChain
	out      *libol.SubLogger
	lock     sync.Mutex
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
	g.lock.Lock()
	defer g.lock.Unlock()

	g.sources[source] = source
}

func (g *ZGuest) HasSource(source string) bool {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.sources[source]; ok {
		return true
	}
	return false
}

func (g *ZGuest) DelSource(source string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.sources[source]; ok {
		delete(g.sources, source)
	}
}

func (g *ZGuest) addRuleX(rule cn.IpRule) {
	if err := g.chain.AddRuleX(rule); err != nil {
		g.out.Warn("ZTrust.AddRuleX: %s", err)
	}
}

func (g *ZGuest) delRuleX(rule cn.IpRule) {
	if err := g.chain.DelRuleX(rule); err != nil {
		g.out.Warn("ZTrust.DelRuleX: %s", err)
	}
}

func (g *ZGuest) AddRule(rule *KnockRule) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if dst, ok := g.rules[rule.Id()]; !ok {
		g.addRuleX(rule.Rule())
		g.rules[rule.Id()] = rule
	} else {
		dst.age = rule.age
		dst.createAt = rule.createAt
	}
}

func (g *ZGuest) DelRule(rule *KnockRule) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.rules[rule.Id()]; ok {
		g.delRuleX(rule.Rule())
		delete(g.rules, rule.Id())
	}
}

func (g *ZGuest) Stop() {
	g.chain.Cancel()
}

func (g *ZGuest) Clear() {
	g.lock.Lock()
	defer g.lock.Unlock()

	removed := make([]*KnockRule, 0, 32)
	for _, rule := range g.rules {
		if rule.Expire() {
			removed = append(removed, rule)
		}
	}
	for _, rule := range removed {
		g.out.Info("ZTrust.Clear: %s", rule.Id())
		delete(g.rules, rule.Id())
		g.delRuleX(rule.Rule())
	}
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
		Comment: "ZTrust Deny All",
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
		createAt:    time.Now(),
		age:         int64(age),
	})
	return nil
}

func (z *ZTrust) Update() {
	for {
		for _, guest := range z.guests {
			guest.Clear()
		}

		time.Sleep(time.Second * 3)
	}
}

func (z *ZTrust) addRuleX(rule cn.IpRule) {
	if err := z.chain.AddRuleX(rule); err != nil {
		z.out.Warn("ZTrust.AddRuleX: %s", err)
	}
}

func (z *ZTrust) delRuleX(rule cn.IpRule) {
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
	if !guest.HasSource(source) {
		z.addRuleX(cn.IpRule{
			Source:  source,
			Comment: "User " + guest.username + "@" + guest.network,
			Jump:    guest.Chain(),
			Order:   "-I",
		})
	}
	guest.AddSource(source)
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
	if guest.HasSource(source) {
		z.delRuleX(cn.IpRule{
			Source:  source,
			Comment: guest.username + "." + guest.network,
			Jump:    guest.Chain(),
		})
	}
	guest.DelSource(source)
	return nil
}

func (z *ZTrust) Start() {
	z.out.Info("ZTrust.Start")
	z.chain.Install()
	libol.Go(z.Update)
}

func (z *ZTrust) Stop() {
	z.out.Info("ZTrust.Stop")
	z.chain.Cancel()
	for _, guest := range z.guests {
		guest.Stop()
	}
}

func (z *ZTrust) ListGuest(call func(obj schema.ZGuest)) {
	for _, guest := range z.guests {
		for _, source := range guest.sources {
			obj := schema.ZGuest{
				Name:    guest.username,
				Network: guest.network,
				Address: source,
			}
			call(obj)
		}
	}
}

func (z *ZTrust) ListKnock(name string, call func(obj schema.KnockRule)) {
	guest, ok := z.guests[name]
	if !ok {
		return
	}

	now := time.Now()
	for _, rule := range guest.rules {
		createAt := rule.createAt
		obj := schema.KnockRule{
			Name:     name,
			Network:  z.network,
			Protocol: rule.protocol,
			Dest:     rule.destination,
			Port:     rule.port,
			CreateAt: createAt.Unix(),
			Age:      int(rule.age + createAt.Unix() - now.Unix()),
		}
		call(obj)
	}
}
