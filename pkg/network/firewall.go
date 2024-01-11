package network

import (
	"sync"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/moby/libnetwork/iptables"
)

const (
	OLCInput   = "XTT_in"
	OLCForward = "XTT_for"
	OLCOutput  = "XTT_out"
	OLCPre     = "XTT_pre"
	OLCPost    = "XTT_pos"
)

type FireWallGlobal struct {
	lock   sync.Mutex
	chains IPChains
	rules  IPRules
}

func NewFireWallGlobal(flows []config.FlowRule) *FireWallGlobal {
	f := &FireWallGlobal{
		chains: make(IPChains, 0, 8),
		rules:  make(IPRules, 0, 32),
	}
	// Load custom rules.
	for _, rule := range flows {
		f.rules = f.rules.Add(IPRule{
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
			CtState:  rule.CtState,
		})
	}
	return f
}

func (f *FireWallGlobal) addOLC() {
	f.AddChain(IPChain{Table: TFilter, Name: OLCInput})
	f.AddChain(IPChain{Table: TFilter, Name: OLCForward})
	f.AddChain(IPChain{Table: TFilter, Name: OLCOutput})
	f.AddChain(IPChain{Table: TNat, Name: OLCPre})
	f.AddChain(IPChain{Table: TNat, Name: OLCInput})
	f.AddChain(IPChain{Table: TNat, Name: OLCPost})
	f.AddChain(IPChain{Table: TNat, Name: OLCOutput})
	f.AddChain(IPChain{Table: TMangle, Name: OLCPre})
	f.AddChain(IPChain{Table: TMangle, Name: OLCInput})
	f.AddChain(IPChain{Table: TMangle, Name: OLCForward})
	f.AddChain(IPChain{Table: TMangle, Name: OLCPost})
	f.AddChain(IPChain{Table: TMangle, Name: OLCOutput})
	f.AddChain(IPChain{Table: TRaw, Name: OLCPre})
	f.AddChain(IPChain{Table: TRaw, Name: OLCOutput})
}

func (f *FireWallGlobal) jumpOLC() {
	// Filter Table
	f.AddRule(IPRule{Order: "-I", Table: TFilter, Chain: CInput, Jump: OLCInput})
	f.AddRule(IPRule{Order: "-I", Table: TFilter, Chain: CForward, Jump: OLCForward})
	f.AddRule(IPRule{Order: "-I", Table: TFilter, Chain: COutput, Jump: OLCOutput})

	// NAT Table
	f.AddRule(IPRule{Order: "-I", Table: TNat, Chain: CPre, Jump: OLCPre})
	f.AddRule(IPRule{Order: "-I", Table: TNat, Chain: CInput, Jump: OLCInput})
	f.AddRule(IPRule{Order: "-I", Table: TNat, Chain: CPost, Jump: OLCPost})
	f.AddRule(IPRule{Order: "-I", Table: TNat, Chain: COutput, Jump: OLCOutput})

	// Mangle Table
	f.AddRule(IPRule{Order: "-I", Table: TMangle, Chain: CPre, Jump: OLCPre})
	f.AddRule(IPRule{Order: "-I", Table: TMangle, Chain: CInput, Jump: OLCInput})
	f.AddRule(IPRule{Order: "-I", Table: TMangle, Chain: CForward, Jump: OLCForward})
	f.AddRule(IPRule{Order: "-I", Table: TMangle, Chain: CPost, Jump: OLCPost})
	f.AddRule(IPRule{Order: "-I", Table: TMangle, Chain: COutput, Jump: OLCOutput})

	// Raw Table
	f.AddRule(IPRule{Order: "-I", Table: TRaw, Chain: CPre, Jump: OLCPre})
	f.AddRule(IPRule{Order: "-I", Table: TRaw, Chain: COutput, Jump: OLCOutput})
}

func (f *FireWallGlobal) Initialize() {
	IPTableInit()
	// Init chains
	f.addOLC()
	f.jumpOLC()
}

func (f *FireWallGlobal) AddChain(chain IPChain) {
	f.chains = f.chains.Add(chain)
}

func (f *FireWallGlobal) AddRule(rule IPRule) {
	f.rules = f.rules.Add(rule)
}

func (f *FireWallGlobal) InstallRule(rule IPRule) error {
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

func (f *FireWallGlobal) install() {
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

func (f *FireWallGlobal) Start() {
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

func (f *FireWallGlobal) cancel() {
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

func (f *FireWallGlobal) CancelRule(rule IPRule) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, err := rule.Opr("-D"); err != nil {
		return err
	}
	f.rules = f.rules.Remove(rule)
	return nil
}

func (f *FireWallGlobal) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()
	libol.Info("FireWall.Stop")
	f.cancel()
}

func (f *FireWallGlobal) Refresh() {
	f.cancel()
	f.install()
}

type FireWallChain struct {
	lock   sync.Mutex
	name   string
	parent string
	rules  IPRules
	table  string
}

func NewFireWallChain(name, table, parent string) *FireWallChain {
	return &FireWallChain{
		name:   name,
		table:  table,
		parent: parent,
	}
}

func (ch *FireWallChain) Chain() IPChain {
	name := ch.name
	if ch.parent != "" {
		name = ch.parent + "-" + ch.name
	}
	if len(name) > 28 {
		name = name[:28]
	}
	return IPChain{
		Table: ch.table,
		Name:  name,
		From:  ch.parent,
	}
}

func (ch *FireWallChain) Jump() IPRule {
	c := ch.Chain()
	return IPRule{
		Order: "-I",
		Table: c.Table,
		Chain: c.From,
		Jump:  c.Name,
	}
}

func (ch *FireWallChain) AddRuleX(rule IPRule) error {
	chain := ch.Chain()
	rule.Table = chain.Table
	rule.Chain = chain.Name
	order := rule.Order
	if order == "" {
		order = "-A"
	}
	if _, err := rule.Opr(order); err != nil {
		return err
	}
	return nil
}

func (ch *FireWallChain) DelRuleX(rule IPRule) error {
	chain := ch.Chain()
	rule.Table = chain.Table
	rule.Chain = chain.Name
	if _, err := rule.Opr("-D"); err != nil {
		return err
	}
	return nil
}

func (ch *FireWallChain) AddRule(rule IPRule) {
	chain := ch.Chain()
	rule.Table = chain.Table
	rule.Chain = chain.Name
	ch.rules = ch.rules.Add(rule)
}

func (ch *FireWallChain) Install() {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	c := ch.Chain()
	if _, err := c.Opr("-N"); err != nil {
		libol.Error("FireWallChain.new %s", err)
	}

	for _, r := range ch.rules {
		order := r.Order
		if order == "" {
			order = "-A"
		}
		if _, err := r.Opr(order); err != nil {
			libol.Error("FireWallChain.install %s", err)
		}
	}

	j := ch.Jump()
	if j.Chain != "" {
		if _, err := j.Opr(j.Order); err != nil {
			libol.Error("FireWallChain.new %s", err)
		}
	}
}

func (ch *FireWallChain) Cancel() {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	j := ch.Jump()
	if j.Chain != "" {
		if _, err := j.Opr("-D"); err != nil {
			libol.Error("FireWallChain.cancel %s", err)
		}
	}
	for _, r := range ch.rules {
		if _, err := r.Opr("-D"); err != nil {
			libol.Warn("FireWall.cancel %s", err)
		}
	}

	c := ch.Chain()
	if _, err := c.Opr("-X"); err != nil {
		libol.Error("FireWallChain.free %s", err)
	}
}

type FireWallFilter struct {
	name string
	In   *FireWallChain
	Out  *FireWallChain
	For  *FireWallChain
}

func NewFireWallFilter(name string) *FireWallFilter {
	return &FireWallFilter{
		In:  NewFireWallChain(name, TFilter, OLCInput),
		For: NewFireWallChain(name, TFilter, OLCForward),
		Out: NewFireWallChain(name, TFilter, OLCOutput),
	}
}

func (f *FireWallFilter) Install() {
	// Install Chain Rules
	f.In.Install()
	f.For.Install()
	f.Out.Install()
}

func (f *FireWallFilter) Cancel() {
	// Cancel Chain Rules
	f.In.Cancel()
	f.For.Cancel()
	f.Out.Cancel()
}

type FireWallNATPre struct {
	*FireWallChain
}

func (ch *FireWallNATPre) Chain() IPChain {
	return IPChain{
		Table: TNat,
		Name:  OLCPre + "-" + ch.name,
		From:  ch.parent,
	}
}

type FireWallNAT struct {
	name string
	Pre  *FireWallChain
	In   *FireWallChain
	Out  *FireWallChain
	Post *FireWallChain
}

func NewFireWallNAT(name string) *FireWallNAT {
	return &FireWallNAT{
		Pre:  NewFireWallChain(name, TNat, OLCPre),
		In:   NewFireWallChain(name, TNat, OLCInput),
		Out:  NewFireWallChain(name, TNat, OLCOutput),
		Post: NewFireWallChain(name, TNat, OLCPost),
	}
}

func (n *FireWallNAT) Install() {
	// Install Chain Rules
	n.Pre.Install()
	n.In.Install()
	n.Out.Install()
	n.Post.Install()
}

func (n *FireWallNAT) Cancel() {
	// Cancel Chain Rules
	n.Pre.Cancel()
	n.In.Cancel()
	n.Out.Cancel()
	n.Post.Cancel()
}

type FireWallMangle struct {
	name string
	Pre  *FireWallChain
	In   *FireWallChain
	For  *FireWallChain
	Out  *FireWallChain
	Post *FireWallChain
}

func NewFireWallMangle(name string) *FireWallMangle {
	return &FireWallMangle{
		Pre:  NewFireWallChain(name, TMangle, OLCPre),
		In:   NewFireWallChain(name, TMangle, OLCInput),
		For:  NewFireWallChain(name, TMangle, OLCForward),
		Out:  NewFireWallChain(name, TMangle, OLCOutput),
		Post: NewFireWallChain(name, TMangle, OLCPost),
	}
}

func (m *FireWallMangle) Install() {
	// Install Chain Rules
	m.Pre.Install()
	m.In.Install()
	m.For.Install()
	m.Out.Install()
	m.Post.Install()
}

func (m *FireWallMangle) Cancel() {
	// Cancel Chain Rules
	m.Pre.Cancel()
	m.In.Cancel()
	m.For.Cancel()
	m.Out.Cancel()
	m.Post.Cancel()
}

type FireWallRaw struct {
	name string
	Pre  *FireWallChain
	Out  *FireWallChain
}

func NewFireWallRaw(name string) *FireWallRaw {
	return &FireWallRaw{
		Pre: NewFireWallChain(name, TRaw, OLCPre),
		Out: NewFireWallChain(name, TRaw, OLCOutput),
	}
}
func (r *FireWallRaw) Install() {
	// Install Chain Rules
	r.Pre.Install()
	r.Out.Install()
}

func (r *FireWallRaw) Cancel() {
	// Cancel Chain Rules
	r.Pre.Cancel()
	r.Out.Cancel()
}

type FireWallTable struct {
	Filter *FireWallFilter
	Nat    *FireWallNAT
	Mangle *FireWallMangle
	Raw    *FireWallRaw
}

func NewFireWallTable(name string) *FireWallTable {
	IPTableInit()
	return &FireWallTable{
		Filter: NewFireWallFilter(name),
		Nat:    NewFireWallNAT(name),
		Mangle: NewFireWallMangle(name),
		Raw:    NewFireWallRaw(name),
	}
}

func (t *FireWallTable) Start() {
	t.Filter.Install()
	t.Nat.Install()
	t.Mangle.Install()
	t.Raw.Install()
}

func (t *FireWallTable) Stop() {
	t.Raw.Cancel()
	t.Mangle.Cancel()
	t.Nat.Cancel()
	t.Filter.Cancel()
}
