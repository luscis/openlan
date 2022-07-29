package network

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/moby/libnetwork/iptables"
	"runtime"
	"strconv"
	"strings"
)

const (
	TNat     = "nat"
	TRaw     = "raw"
	TMangle  = "mangle"
	TFilter  = "filter"
	CInput   = "INPUT"
	CForward = "FORWARD"
	COutput  = "OUTPUT"
	CPost    = "POSTROUTING"
	CPre     = "PREROUTING"
	CMasq    = "MASQUERADE"
	CNoTrk   = "NOTRACK"
	CSnat    = "SNAT"
)

type IpRule struct {
	Table    string
	Chain    string
	Source   string
	ToSource string
	NoSource string
	Dest     string
	ToDest   string
	NoDest   string
	Proto    string
	DstPort  string
	SrcPort  string
	Input    string
	Output   string
	Comment  string
	Jump     string
	SetMss   int
	Order    string
	Match    string
	TcpFlag  []string
}

type IpRules []IpRule

func (ru IpRule) Itoa(value int) string {
	return strconv.Itoa(value)
}

func (ru IpRule) Args() []string {
	var args []string

	if ru.Source != "" {
		args = append(args, "-s", ru.Source)
	} else if ru.NoSource != "" {
		args = append(args, "!")
		args = append(args, "-s", ru.NoSource)
	}
	if ru.Dest != "" {
		args = append(args, "-d", ru.Dest)
	} else if ru.NoDest != "" {
		args = append(args, "!")
		args = append(args, "-d", ru.NoDest)
	}
	if ru.Proto != "" {
		args = append(args, "-p", ru.Proto)
	}
	if ru.Match != "" {
		args = append(args, "-m", ru.Match)
	}
	if len(ru.TcpFlag) > 0 {
		args = append(args, "--tcp-flags", ru.TcpFlag[0], ru.TcpFlag[1])
	}
	if len(ru.SrcPort) > 0 {
		args = append(args, "--sport", ru.SrcPort)
	}
	if len(ru.DstPort) > 0 {
		if ru.Match == "multiport" {
			args = append(args, "--dports", ru.DstPort)
		} else {
			args = append(args, "--dport", ru.DstPort)
		}
	}
	if ru.Input != "" {
		args = append(args, "-i", ru.Input)
	}
	if ru.Output != "" {
		args = append(args, "-o", ru.Output)
	}
	if ru.Jump != "" {
		jump := strings.ToUpper(ru.Jump)
		if jump == "DROP" || jump == "ACCEPT" {
			args = append(args, "-j", jump)
		} else {
			args = append(args, "-j", ru.Jump)
		}
		if ru.SetMss > 0 {
			args = append(args, "--set-mss", ru.Itoa(ru.SetMss))
		}
	} else {
		args = append(args, "-j", "ACCEPT")
	}
	if ru.ToSource != "" {
		args = append(args, "--to-source", ru.ToSource)
	}
	if ru.ToDest != "" {
		args = append(args, "--to-destination", ru.ToDest)
	}
	return args
}

func (ru IpRule) Exist() bool {
	table := iptables.Table(ru.Table)
	chain := ru.Chain
	args := ru.Args()
	return iptables.Exists(table, chain, args...)
}

func (ru IpRule) String() string {
	return ru.Table + " " + ru.Chain + " " + strings.Join(ru.Args(), " ")
}

func (ru IpRule) Eq(obj IpRule) bool {
	return ru.String() == obj.String()
}

func (ru IpRule) Opr(opr string) ([]byte, error) {
	libol.Debug("IpRuleOpr: %s, %v", opr, ru)
	table := iptables.Table(ru.Table)
	chain := ru.Chain
	switch runtime.GOOS {
	case "linux":
		args := ru.Args()
		fullArgs := append([]string{"-t", ru.Table, opr, ru.Chain}, args...)
		if opr == "-I" || opr == "-A" {
			if iptables.Exists(table, chain, args...) {
				return nil, nil
			}
		}
		return iptables.Raw(fullArgs...)
	default:
		return nil, libol.NewErr("iptables notSupport %s", runtime.GOOS)
	}
}

func (rules IpRules) Add(obj IpRule) IpRules {
	if !rules.Has(obj) {
		return append(rules, obj)
	}
	return rules
}

func (rules IpRules) Has(rule IpRule) bool {
	for _, r := range rules {
		if r.Eq(rule) {
			return true
		}
	}
	return false
}

func (rules IpRules) Remove(obj IpRule) IpRules {
	index := 0
	news := make(IpRules, 0, 32)
	find := false
	for _, item := range rules {
		if !find && item.Eq(obj) {
			find = true
			continue
		}
		news[index] = item
		index++
	}
	return news[:index]
}

type IpChain struct {
	Table string
	Name  string
}

type IpChains []IpChain

func (ch IpChain) Opr(opr string) ([]byte, error) {
	libol.Debug("IpChainOpr: %s, %v", opr, ch)
	table := iptables.Table(ch.Table)
	name := ch.Name
	switch runtime.GOOS {
	case "linux":
		if opr == "-N" {
			if iptables.ExistChain(name, table) {
				return nil, nil
			}
			if _, err := iptables.NewChain(name, table, true); err != nil {
				return nil, err
			}
		} else if opr == "-X" {
			if err := iptables.RemoveExistingChain(name, table); err != nil {
				return nil, err
			}
		}
	default:
		return nil, libol.NewErr("iptables notSupport %s", runtime.GOOS)
	}
	return nil, nil
}

func (ch IpChain) Eq(obj IpChain) bool {
	if ch.Table != obj.Table {
		return false
	}
	if ch.Name != obj.Name {
		return false
	}
	return true
}

func (chains IpChains) Add(obj IpChain) IpChains {
	return append(chains, obj)
}

func (chains IpChains) Pop(obj IpChain) IpChains {
	index := 0
	news := make(IpChains, 0, 32)
	find := false
	for _, item := range chains {
		if !find && item.Eq(obj) {
			find = true
			continue
		}
		news[index] = item
		index++
	}
	return news[:index]
}

func IpInit() {
	if err := iptables.FirewalldInit(); err != nil {
		libol.Error("IpInit %s", err)
	}
}
