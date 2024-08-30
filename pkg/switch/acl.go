package cswitch

import (
	"fmt"
	"strconv"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type ACLRule struct {
	SrcIp   string
	DstIp   string
	Proto   string // TCP, UDP or ICMP
	SrcPort int
	DstPort int
	Action  string // DROP or ACCEPT
	rule    *cn.IPRule
}

func (r *ACLRule) Id() string {
	return fmt.Sprintf("%s:%s:%s:%d:%d", r.SrcIp, r.DstIp, r.Proto, r.DstPort, r.SrcPort)
}

func (r *ACLRule) Rule() cn.IPRule {
	if r.rule == nil {
		r.rule = &cn.IPRule{
			Dest:   r.DstIp,
			Source: r.SrcIp,
			Proto:  r.Proto,
			Jump:   r.Action,
		}
		if r.DstPort > 0 {
			r.rule.DstPort = strconv.Itoa(r.DstPort)
		}
		if r.SrcPort > 0 {
			r.rule.SrcPort = strconv.Itoa(r.SrcPort)
		}
	}
	return *r.rule
}

type ACL struct {
	Name  string
	Rules map[string]*ACLRule
	chain *cn.FireWallChain
	out   *libol.SubLogger
}

func NewACL(name string) *ACL {
	return &ACL{
		Name:  name,
		out:   libol.NewSubLogger(name),
		Rules: make(map[string]*ACLRule, 32),
	}
}

func (a *ACL) Chain() string {
	return "ATT_" + a.Name
}

func (a *ACL) Initialize() {
	a.chain = cn.NewFireWallChain(a.Chain(), cn.TRaw, "")
}

func (a *ACL) Start() {
	a.out.Info("ACL.Start")
	cfg := co.GetAcl(a.Name)
	if cfg != nil {
		for _, rule := range cfg.Rules {
			a.addRule(rule)
		}
	}
	a.chain.Install()
}

func (a *ACL) Stop() {
	a.out.Info("ACL.Stop")
	a.chain.Cancel()
}

func (a *ACL) addRule(rule *co.ACLRule) {
	ar := &ACLRule{
		Proto:   rule.Proto,
		DstIp:   rule.DstIp,
		SrcIp:   rule.SrcIp,
		DstPort: rule.DstPort,
		SrcPort: rule.SrcPort,
		Action:  rule.Action,
	}

	a.out.Info("ACL.addRule %s", ar.Id())

	if _, ok := a.Rules[ar.Id()]; !ok {
		a.chain.AddRule(ar.Rule())
		a.Rules[ar.Id()] = ar
	}
}

func (a *ACL) AddRule(rule *schema.ACLRule) error {
	ar := &ACLRule{
		Proto:   rule.Proto,
		DstIp:   rule.DstIp,
		SrcIp:   rule.SrcIp,
		DstPort: rule.DstPort,
		SrcPort: rule.SrcPort,
		Action:  rule.Action,
	}

	a.out.Info("ACL.AddRule %s", ar.Id())

	if _, ok := a.Rules[ar.Id()]; ok {
		return libol.NewErr("AddRule: already existed")
	}

	if err := a.chain.AddRuleX(ar.Rule()); err == nil {
		a.Rules[ar.Id()] = ar
	} else {
		a.out.Warn("ACL.AddRule %s", err)
	}

	return nil
}

func (a *ACL) DelRule(rule *schema.ACLRule) error {
	ar := &ACLRule{
		Proto:   rule.Proto,
		DstIp:   rule.DstIp,
		SrcIp:   rule.SrcIp,
		DstPort: rule.DstPort,
		SrcPort: rule.SrcPort,
		Action:  rule.Action,
	}

	a.out.Info("ACL.DelRule %s", ar.Id())

	if _, ok := a.Rules[ar.Id()]; !ok {
		return nil
	}

	if err := a.chain.DelRuleX(ar.Rule()); err == nil {
		delete(a.Rules, ar.Id())
	} else {
		a.out.Warn("ACL.DelRule %s", err)
	}

	return nil
}

func (a *ACL) ListRules(call func(obj schema.ACLRule)) {
	for _, rule := range a.Rules {
		obj := schema.ACLRule{
			SrcIp:   rule.SrcIp,
			DstIp:   rule.DstIp,
			SrcPort: rule.SrcPort,
			DstPort: rule.DstPort,
			Proto:   rule.Proto,
			Action:  rule.Action,
		}
		call(obj)
	}
}

func (a *ACL) SaveRule() {
	cfg := co.GetAcl(a.Name)
	cfg.Rules = nil
	for _, rule := range a.Rules {
		cr := &co.ACLRule{
			DstIp:   rule.DstIp,
			SrcIp:   rule.SrcIp,
			Proto:   rule.Proto,
			DstPort: rule.DstPort,
			SrcPort: rule.SrcPort,
			Action:  rule.Action,
		}
		cfg.Rules = append(cfg.Rules, cr)
	}
	cfg.Save()
}
