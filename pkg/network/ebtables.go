package network

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

const (
	TEbFilter  = "filter"
	CEbInput   = "INPUT"
	CEbForward = "FORWARD"
)

type EBRule struct {
	Table      string
	Chain      string
	Source     string
	Dest       string
	Proto      string
	DstPort    string
	SrcPort    string
	Input      string
	LogicalIn  string
	LogicalOut string
	Jump       string
	Order      string
}

func (ru EBRule) Args() []string {
	var args []string

	if ru.Input != "" {
		args = append(args, "-i", ru.Input)
	}
	if ru.LogicalIn != "" {
		args = append(args, "--logical-in", ru.LogicalIn)
	}
	if ru.LogicalOut != "" {
		args = append(args, "--logical-out", ru.LogicalOut)
	}

	ipArgs := make([]string, 0, 10)
	if ru.Source != "" {
		ipArgs = append(ipArgs, "--ip-src", ru.Source)
	}
	if ru.Dest != "" {
		ipArgs = append(ipArgs, "--ip-dst", ru.Dest)
	}
	if ru.Proto != "" {
		ipArgs = append(ipArgs, "--ip-proto", strings.ToLower(ru.Proto))
	}
	if ru.SrcPort != "" {
		ipArgs = append(ipArgs, "--ip-sport", ru.SrcPort)
	}
	if ru.DstPort != "" {
		ipArgs = append(ipArgs, "--ip-dport", ru.DstPort)
	}
	jump := strings.ToUpper(ru.Jump)
	if len(ipArgs) > 0 || jump == "ACCEPT" || jump == "DROP" {
		args = append(args, "-p", "IPv4")
		args = append(args, ipArgs...)
	}

	if ru.Jump != "" {
		if jump == "ACCEPT" || jump == "DROP" {
			args = append(args, "-j", jump)
		} else {
			args = append(args, "-j", ru.Jump)
		}
	} else {
		args = append(args, "-j", "ACCEPT")
	}

	return args
}

func (ru EBRule) String() string {
	order := ru.Order
	if order == "" {
		order = "-A"
	}
	elems := append([]string{"-t", ru.Table, order, ru.Chain}, ru.Args()...)
	return strings.Join(elems, " ")
}

func (ru EBRule) Eq(obj EBRule) bool {
	return ru.String() == obj.String()
}

func (ru EBRule) Opr(opr string) ([]byte, error) {
	libol.Debug("EBRuleOpr: %s, %v", opr, ru)
	switch runtime.GOOS {
	case "linux":
		args := append([]string{"-t", ru.Table, opr, ru.Chain}, ru.Args()...)
		return exec.Command("ebtables", args...).CombinedOutput()
	default:
		return nil, libol.NewErr("ebtables notSupport %s", runtime.GOOS)
	}
}

type EBRules []EBRule

func (rules EBRules) Add(obj EBRule) EBRules {
	if !rules.Has(obj) {
		return append(rules, obj)
	}
	return rules
}

func (rules EBRules) Has(rule EBRule) bool {
	for _, r := range rules {
		if r.Eq(rule) {
			return true
		}
	}
	return false
}

func (rules EBRules) Remove(obj EBRule) EBRules {
	news := make(EBRules, 0, len(rules))
	removed := false
	for _, item := range rules {
		if !removed && item.Eq(obj) {
			removed = true
			continue
		}
		news = append(news, item)
	}
	return news
}

type EBChain struct {
	Table string
	Name  string
}

func (ch EBChain) Opr(opr string) ([]byte, error) {
	switch runtime.GOOS {
	case "linux":
		switch opr {
		case "-N":
			if _, err := exec.Command("ebtables", "-t", ch.Table, "-L", ch.Name).CombinedOutput(); err == nil {
				return nil, nil
			}
			return exec.Command("ebtables", "-t", ch.Table, "-N", ch.Name).CombinedOutput()
		case "-X":
			_, _ = exec.Command("ebtables", "-t", ch.Table, "-F", ch.Name).CombinedOutput()
			return exec.Command("ebtables", "-t", ch.Table, "-X", ch.Name).CombinedOutput()
		case "-F":
			return exec.Command("ebtables", "-t", ch.Table, "-F", ch.Name).CombinedOutput()
		}
	default:
		return nil, libol.NewErr("ebtables notSupport %s", runtime.GOOS)
	}
	return nil, nil
}

type EBFireWallChain struct {
	name  string
	table string
	ready bool
	rules EBRules
	hooks EBRules
}

func NewEBFireWallChain(name, table string) *EBFireWallChain {
	return &EBFireWallChain{
		name:  name,
		table: table,
		rules: make(EBRules, 0, 32),
		hooks: make(EBRules, 0, 8),
	}
}

func (ch *EBFireWallChain) Chain() EBChain {
	name := ch.name
	if len(name) > 28 {
		name = name[:28]
	}
	return EBChain{
		Table: ch.table,
		Name:  name,
	}
}

func (ch *EBFireWallChain) Prepare() {
	if ch.ready {
		return
	}
	c := ch.Chain()
	if _, err := c.Opr("-N"); err != nil {
		libol.Error("EBFireWallChain.Prepare %s", err)
	}
	ch.ready = true
}

func (ch *EBFireWallChain) AddRuleX(rule EBRule) error {
	ch.Prepare()
	chain := ch.Chain()
	rule.Table = chain.Table
	rule.Chain = chain.Name
	order := rule.Order
	if order == "" {
		order = "-A"
	}
	if order != "-A" && order != "-I" {
		order = "-A"
	}
	if _, err := rule.Opr(order); err != nil {
		libol.Error("EBFireWallChain.AddRuleX %s", err)
		return err
	}
	return nil
}

func (ch *EBFireWallChain) DelRuleX(rule EBRule) error {
	chain := ch.Chain()
	rule.Table = chain.Table
	rule.Chain = chain.Name
	if _, err := rule.Opr("-D"); err != nil {
		libol.Warn("EBFireWallChain.DelRuleX %s", err)
		return err
	}
	return nil
}

func (ch *EBFireWallChain) AddRule(rule EBRule) {
	chain := ch.Chain()
	rule.Table = chain.Table
	rule.Chain = chain.Name
	ch.rules = ch.rules.Add(rule)
}

func (ch *EBFireWallChain) Install() {
	ch.Prepare()
	for _, r := range ch.rules {
		order := r.Order
		if order == "" {
			order = "-A"
		}
		if order != "-A" && order != "-I" {
			order = "-A"
		}
		if _, err := r.Opr(order); err != nil {
			libol.Error("EBFireWallChain.Install %s", err)
		}
	}
	for _, hook := range ch.hooks {
		if _, err := hook.Opr("-I"); err != nil {
			libol.Error("EBFireWallChain.InstallHook %s", err)
		}
	}
}

func (ch *EBFireWallChain) AddHook(rule EBRule) error {
	ch.Prepare()
	if ch.hooks.Has(rule) {
		return nil
	}
	if _, err := rule.Opr("-I"); err != nil {
		libol.Error("EBFireWallChain.AddHook %s", err)
		return err
	}
	ch.hooks = ch.hooks.Add(rule)
	return nil
}

func (ch *EBFireWallChain) DelHook(rule EBRule) error {
	if _, err := rule.Opr("-D"); err != nil {
		libol.Warn("EBFireWallChain.DelHook %s", err)
		return err
	}
	ch.hooks = ch.hooks.Remove(rule)
	return nil
}

func (ch *EBFireWallChain) Cancel() {
	for _, hook := range ch.hooks {
		_, _ = hook.Opr("-D")
	}
	ch.hooks = make(EBRules, 0, 8)

	c := ch.Chain()
	if _, err := c.Opr("-X"); err != nil {
		libol.Error("EBFireWallChain.Cancel %s", err)
	}
	ch.ready = false
}

func (ch *EBFireWallChain) Flush() {
	c := ch.Chain()
	if _, err := c.Opr("-F"); err != nil {
		libol.Error("EBFireWallChain.Flush %s", err)
	}
	ch.rules = make(EBRules, 0, 32)
}

func ItoEBPort(value int) string {
	return strconv.Itoa(value)
}
