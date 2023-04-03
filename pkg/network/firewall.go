package network

import (
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/moby/libnetwork/iptables"
	"sync"
)

const (
	OLCInput   = "OPENLAN-IN"
	OLCForward = "OPENLAN-FWD"
	OLCOutput  = "OPENLAN-OUT"
	OLCPre     = "OPENLAN-PRE"
	OLCPost    = "OPENLAN-POST"
)

type FireWall struct {
	lock   sync.Mutex
	chains IpChains
	rules  IpRules
}

func NewFireWall(flows []config.FlowRule) *FireWall {
	f := &FireWall{
		chains: make(IpChains, 0, 8),
		rules:  make(IpRules, 0, 32),
	}
	// Load custom rules.
	for _, rule := range flows {
		f.rules = f.rules.Add(IpRule{
			Table:    rule.Table,
			Chain:    rule.Chain,
			Source:   rule.Source,
			Dest:     rule.Dest,
			Jump:     rule.Jump,
			ToSource: rule.ToSource,
			ToDest:   rule.ToDest,
			Comment:  rule.Comment,
			Proto:    rule.Proto,
			Match:    rule.Match,
			DstPort:  rule.DstPort,
			SrcPort:  rule.SrcPort,
			Input:    rule.Input,
			Output:   rule.Output,
		})
	}
	return f
}

func (f *FireWall) addOLC() {
	f.AddChain(IpChain{Table: TFilter, Name: OLCInput})
	f.AddChain(IpChain{Table: TFilter, Name: OLCForward})
	f.AddChain(IpChain{Table: TFilter, Name: OLCOutput})
	f.AddChain(IpChain{Table: TNat, Name: OLCPre})
	f.AddChain(IpChain{Table: TNat, Name: OLCInput})
	f.AddChain(IpChain{Table: TNat, Name: OLCPost})
	f.AddChain(IpChain{Table: TNat, Name: OLCOutput})
	f.AddChain(IpChain{Table: TMangle, Name: OLCPre})
	f.AddChain(IpChain{Table: TMangle, Name: OLCInput})
	f.AddChain(IpChain{Table: TMangle, Name: OLCForward})
	f.AddChain(IpChain{Table: TMangle, Name: OLCPost})
	f.AddChain(IpChain{Table: TMangle, Name: OLCOutput})
	f.AddChain(IpChain{Table: TRaw, Name: OLCPre})
	f.AddChain(IpChain{Table: TRaw, Name: OLCOutput})
}

func (f *FireWall) jumpOLC() {
	f.AddRule(IpRule{Order: "-I", Table: TFilter, Chain: CInput, Jump: OLCInput})
	f.AddRule(IpRule{Order: "-I", Table: TFilter, Chain: CForward, Jump: OLCForward})
	f.AddRule(IpRule{Order: "-I", Table: TFilter, Chain: COutput, Jump: OLCOutput})
	f.AddRule(IpRule{Order: "-I", Table: TNat, Chain: CPre, Jump: OLCPre})
	f.AddRule(IpRule{Order: "-I", Table: TNat, Chain: CInput, Jump: OLCInput})
	f.AddRule(IpRule{Order: "-I", Table: TNat, Chain: CPost, Jump: OLCPost})
	f.AddRule(IpRule{Order: "-I", Table: TNat, Chain: COutput, Jump: OLCOutput})
	f.AddRule(IpRule{Order: "-I", Table: TMangle, Chain: CPre, Jump: OLCPre})
	f.AddRule(IpRule{Order: "-I", Table: TMangle, Chain: CInput, Jump: OLCInput})
	f.AddRule(IpRule{Order: "-I", Table: TMangle, Chain: CForward, Jump: OLCForward})
	f.AddRule(IpRule{Order: "-I", Table: TMangle, Chain: CPost, Jump: OLCPost})
	f.AddRule(IpRule{Order: "-I", Table: TMangle, Chain: COutput, Jump: OLCOutput})
	f.AddRule(IpRule{Order: "-I", Table: TRaw, Chain: CPre, Jump: OLCPre})
	f.AddRule(IpRule{Order: "-I", Table: TRaw, Chain: COutput, Jump: OLCOutput})
}

func (f *FireWall) Initialize() {
	IptInit()
	// Init chains
	f.addOLC()
	f.jumpOLC()
}

func (f *FireWall) AddChain(chain IpChain) {
	f.chains = f.chains.Add(chain)
}

func (f *FireWall) AddRule(rule IpRule) {
	f.rules = f.rules.Add(rule)
}

func (f *FireWall) InstallRule(rule IpRule) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	order := rule.Order
	if order == "" {
		order = "-A"
	}
	if _, err := rule.Opr(order); err != nil {
		return err
	}
	f.rules = f.rules.Add(rule)
	return nil
}

func (f *FireWall) install() {
	for _, c := range f.chains {
		if _, err := c.Opr("-N"); err != nil {
			libol.Error("FireWall.install %s", err)
		}
	}
	for _, r := range f.rules {
		order := r.Order
		if order == "" {
			order = "-A"
		}
		if _, err := r.Opr(order); err != nil {
			libol.Error("FireWall.install %s", err)
		}
	}
}

func (f *FireWall) Start() {
	f.lock.Lock()
	defer f.lock.Unlock()
	libol.Info("FireWall.Start")
	f.install()
	iptables.OnReloaded(func() {
		libol.Info("FireWall.Start OnReloaded")
		f.lock.Lock()
		defer f.lock.Unlock()
		f.install()
	})
}

func (f *FireWall) cancel() {
	for _, r := range f.rules {
		if _, err := r.Opr("-D"); err != nil {
			libol.Warn("FireWall.cancel %s", err)
		}
	}
	for _, c := range f.chains {
		if _, err := c.Opr("-X"); err != nil {
			libol.Warn("FireWall.cancel %s", err)
		}
	}
}

func (f *FireWall) CancelRule(rule IpRule) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, err := rule.Opr("-D"); err != nil {
		return err
	}
	f.rules = f.rules.Remove(rule)
	return nil
}

func (f *FireWall) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()
	libol.Info("FireWall.Stop")
	f.cancel()
}

func (f *FireWall) Refresh() {
	f.cancel()
	f.install()
}
