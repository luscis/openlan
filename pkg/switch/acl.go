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
	ipRule  *cn.IPRule
	ebRule  *cn.EBRule
}

func (r *ACLRule) Id() string {
	return fmt.Sprintf("%s %s:%s:%s:%d:%d", r.Action, r.SrcIp, r.DstIp, r.Proto, r.DstPort, r.SrcPort)
}

func (r *ACLRule) ToIPRule() cn.IPRule {
	if r.ipRule == nil {
		r.ipRule = &cn.IPRule{
			Dest:   r.DstIp,
			Source: r.SrcIp,
			Proto:  r.Proto,
			Jump:   r.Action,
			Order:  "-I",
		}
		if r.DstPort > 0 {
			r.ipRule.DstPort = strconv.Itoa(r.DstPort)
		}
		if r.SrcPort > 0 {
			r.ipRule.SrcPort = strconv.Itoa(r.SrcPort)
		}
	}
	return *r.ipRule
}

func (r *ACLRule) ToEBRule() cn.EBRule {
	if r.ebRule == nil {
		r.ebRule = &cn.EBRule{
			Dest:   r.DstIp,
			Source: r.SrcIp,
			Proto:  r.Proto,
			Jump:   r.Action,
			Order:  "-I",
		}
		if r.DstPort > 0 {
			r.ebRule.DstPort = strconv.Itoa(r.DstPort)
		}
		if r.SrcPort > 0 {
			r.ebRule.SrcPort = strconv.Itoa(r.SrcPort)
		}
	}
	return *r.ebRule
}

type ACL struct {
	Name    string
	Rules   map[string]*ACLRule
	ipchain *cn.FireWallChain
	ebchain *cn.EBFireWallChain
	out     *libol.SubLogger
}

func NewACL(name string) *ACL {
	return &ACL{
		Name:  name,
		out:   libol.NewSubLogger(name),
		Rules: make(map[string]*ACLRule, 32),
	}
}

func (a *ACL) Chain() string {
	return "AT_" + a.Name
}

func (a *ACL) Initialize() {
	a.ipchain = cn.NewFireWallChain(a.Chain(), cn.TRaw, "")
	a.ebchain = cn.NewEBFireWallChain(a.Chain(), cn.TEbFilter)
}

func (a *ACL) Start() {
	a.out.Info("ACL.Start")
	cfg := co.GetAcl(a.Name)
	if cfg != nil {
		for _, rule := range cfg.Rules {
			ar := &ACLRule{
				Proto:   rule.Proto,
				DstIp:   rule.DstIp,
				SrcIp:   rule.SrcIp,
				DstPort: rule.DstPort,
				SrcPort: rule.SrcPort,
				Action:  rule.Action,
			}
			a.addRule(ar)
		}
	}
	a.ipchain.Install()
	a.ebchain.Install()
}

func (a *ACL) Stop() {
	a.out.Info("ACL.Stop")
	a.ipchain.Cancel()
	a.ebchain.Cancel()
}

func (a *ACL) AddInput(input string) {
	a.out.Info("ACL.AddInput %s", input)
	chain := a.ebchain.Chain()
	hooks := []cn.EBRule{
		{
			Table:     chain.Table,
			Chain:     cn.CEbInput,
			LogicalIn: input,
			Jump:      chain.Name,
		},
		{
			Table:     chain.Table,
			Chain:     cn.CEbForward,
			LogicalIn: input,
			Jump:      chain.Name,
		},
	}
	for _, hook := range hooks {
		if err := a.ebchain.AddHook(hook); err != nil {
			a.out.Warn("ACL.AddInput %s %s", input, err)
		}
	}
}

func (a *ACL) DelInput(input string) {
	a.out.Info("ACL.DelInput %s", input)
	chain := a.ebchain.Chain()
	hooks := []cn.EBRule{
		{
			Table:     chain.Table,
			Chain:     cn.CEbInput,
			LogicalIn: input,
			Jump:      chain.Name,
		},
		{
			Table:     chain.Table,
			Chain:     cn.CEbForward,
			LogicalIn: input,
			Jump:      chain.Name,
		},
	}
	for _, hook := range hooks {
		if err := a.ebchain.DelHook(hook); err != nil {
			a.out.Warn("ACL.DelInput %s %s", input, err)
		}
	}
}

func (a *ACL) addRule(ar *ACLRule) {
	a.out.Info("ACL.addRule %s", ar.Id())

	if _, ok := a.Rules[ar.Id()]; ok {
		return
	}

	a.ipchain.AddRuleX(ar.ToIPRule())
	a.ebchain.AddRuleX(ar.ToEBRule())
	a.Rules[ar.Id()] = ar
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
	if _, ok := a.Rules[ar.Id()]; ok {
		return libol.NewErr("AddRule: already existed")
	}

	a.addRule(ar)

	return nil
}

func (a *ACL) delRule(ar *ACLRule) {
	if _, ok := a.Rules[ar.Id()]; !ok {
		return
	}

	if err := a.ipchain.DelRuleX(ar.ToIPRule()); err != nil {
		a.out.Warn("ACL.DelRule %s", err)
	}
	if err := a.ebchain.DelRuleX(ar.ToEBRule()); err != nil {
		a.out.Warn("ACL.DelRule.eb %s", err)
	}
	delete(a.Rules, ar.Id())
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

	if _, ok := a.Rules[ar.Id()]; !ok {
		return nil
	}

	a.delRule(ar)

	return nil
}

func (a *ACL) FlushRules() {
	a.out.Info("ACL.FlushRules")
	a.ipchain.Flush()
	a.ebchain.Flush()
	a.Rules = make(map[string]*ACLRule, 32)
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
