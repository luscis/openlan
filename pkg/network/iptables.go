package network

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/moby/libnetwork/iptables"
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
	CMark    = "MARK"
	CCT      = "CT"
	CNoTrk   = "NOTRACK"
	CSnat    = "SNAT"
	CTcpMss  = "TCPMSS"
)

type IPRule struct {
	Table      string
	Chain      string
	Source     string
	SrcSet     string
	ToSource   string
	NoSource   string
	NoSrcSet   string
	Dest       string
	DestSet    string
	ToDest     string
	NoDest     string
	NoDestSet  string
	Proto      string
	DstPort    string
	SrcPort    string
	Input      string
	Output     string
	Comment    string
	Jump       string
	Limit      string
	LimitBurst string
	SetMss     int
	Mark       uint32
	SetMark    uint32
	Zone       uint32
	Order      string
	Match      string
	CtState    string
	TcpFlag    []string
}

type IPRules []IPRule

func (ru IPRule) Itoa(value int) string {
	return fmt.Sprintf("%d", value)
}

func (ru IPRule) Utoa(value uint32) string {
	return fmt.Sprintf("%d", value)
}

func (ru IPRule) Args() []string {
	var args []string

	if ru.Mark > 0 {
		args = append(args, "-m", "mark", "--mark", ru.Utoa(ru.Mark))
	}

	if ru.Source != "" {
		args = append(args, "-s", ru.Source)
	} else if ru.NoSource != "" {
		args = append(args, "!")
		args = append(args, "-s", ru.NoSource)
	} else if ru.SrcSet != "" {
		args = append(args, "-m", "set", "--match-set", ru.SrcSet, "src")
	} else if ru.NoSrcSet != "" {
		args = append(args, "!")
		args = append(args, "-m", "set", "--match-set", ru.NoSrcSet, "src")
	}

	if ru.Dest != "" {
		args = append(args, "-d", ru.Dest)
	} else if ru.NoDest != "" {
		args = append(args, "!")
		args = append(args, "-d", ru.NoDest)
	} else if ru.DestSet != "" {
		args = append(args, "-m", "set", "--match-set", ru.DestSet, "dst")
	} else if ru.NoDestSet != "" {
		args = append(args, "!")
		args = append(args, "-m", "set", "--match-set", ru.NoDestSet, "dst")
	}

	if ru.CtState != "" {
		args = append(args, "-m", "conntrack", "--ctstate", ru.CtState)
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
	if ru.SrcPort != "" {
		args = append(args, "--sport", ru.SrcPort)
	}
	if ru.DstPort != "" {
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
	if ru.Limit != "" {
		args = append(args, "-m", "limit", "--limit", ru.Limit)
	}
	if ru.LimitBurst != "" {
		args = append(args, "--limit-burst", ru.LimitBurst)
	}
	if ru.Comment != "" {
		args = append(args, "-m", "comment", "--comment", ru.Comment)
	}

	if ru.Jump != "" {
		jump := strings.ToUpper(ru.Jump)
		if jump == "ACCEPT" || jump == "DROP" {
			args = append(args, "-j", jump)
		} else if jump == "MARK" {
			args = append(args, "-j", "MARK", "--set-mark", ru.Utoa(ru.SetMark))
		} else if jump == "CT" {
			args = append(args, "-j", "CT")
			if ru.Zone > 0 {
				args = append(args, "--zone", ru.Utoa(ru.Zone))
			}
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

func (ru IPRule) Exist() bool {
	table := iptables.Table(ru.Table)
	chain := ru.Chain
	args := ru.Args()
	return iptables.Exists(table, chain, args...)
}

func (ru IPRule) String() string {
	elems := append([]string{"-t", ru.Table, "-A", ru.Chain}, ru.Args()...)
	return strings.Join(elems, " ")
}

func (ru IPRule) Eq(obj IPRule) bool {
	return ru.String() == obj.String()
}

func (ru IPRule) Opr(opr string) ([]byte, error) {
	libol.Debug("IPRuleOpr: %s, %v", opr, ru)
	switch runtime.GOOS {
	case "linux":
		if opr == "-A" || opr == "-I" {
			if ru.Exist() {
				return nil, nil
			}
		}
		args := ru.Args()
		fullArgs := append([]string{"-t", ru.Table, opr, ru.Chain}, args...)
		return iptables.Raw(fullArgs...)
	default:
		return nil, libol.NewErr("iptables notSupport %s", runtime.GOOS)
	}
}

func (rules IPRules) Add(obj IPRule) IPRules {
	if !rules.Has(obj) {
		return append(rules, obj)
	}
	return rules
}

func (rules IPRules) Has(rule IPRule) bool {
	for _, r := range rules {
		if r.Eq(rule) {
			return true
		}
	}
	return false
}

func (rules IPRules) Remove(obj IPRule) IPRules {
	index := 0
	news := make(IPRules, 0, 32)
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

type IPChain struct {
	Table string
	Name  string
	From  string
}

type IPChains []IPChain

func (ch IPChain) Opr(opr string) ([]byte, error) {
	table := iptables.Table(ch.Table)
	name := ch.Name
	switch runtime.GOOS {
	case "linux":
		switch opr {
		case "-N":
			if iptables.ExistChain(name, table) {
				return nil, nil
			}
			if _, err := iptables.NewChain(name, table, true); err != nil {
				return nil, err
			}
		case "-X":
			if err := iptables.RemoveExistingChain(name, table); err != nil {
				return nil, err
			}
		case "-F":
			iptables.Raw("-t", ch.Table, "-F", ch.Name)
		}
	default:
		return nil, libol.NewErr("iptables notSupport %s", runtime.GOOS)
	}
	return nil, nil
}

func (ch IPChain) Eq(obj IPChain) bool {
	if ch.Table != obj.Table {
		return false
	}
	if ch.Name != obj.Name {
		return false
	}
	return true
}

func (chains IPChains) Add(obj IPChain) IPChains {
	return append(chains, obj)
}

func (chains IPChains) Pop(obj IPChain) IPChains {
	index := 0
	news := make(IPChains, 0, 32)
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

var __iptablesInit__ = false

func IPTableInit() {
	if __iptablesInit__ {
		return
	}
	__iptablesInit__ = true
	if err := iptables.FirewalldInit(); err != nil {
		libol.Debug("IptInit %s", err)
	}
}
