package config

import "github.com/luscis/openlan/pkg/libol"

type ACL struct {
	File  string     `json:"-" yaml:"-"`
	Name  string     `json:"name" yaml:"name"`
	Rules []*ACLRule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

func (ru *ACL) Save() {
	if err := libol.MarshalSave(ru, ru.File, true); err != nil {
		libol.Error("Switch.Save.Acl %s %s", ru.Name, err)
	}
}

func (ru *ACL) Correct(sw *Switch) {
	for _, rule := range ru.Rules {
		rule.Correct()
	}
	if ru.File == "" {
		ru.File = sw.Dir("acl", ru.Name)
	}
}

type ACLRule struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	SrcIp   string `json:"source,omitempty" yaml:"source,omitempty"`
	DstIp   string `json:"destination,omitempty" yaml:"destination,omitempty"`
	Proto   string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	SrcPort int    `json:"sport,omitempty" yaml:"sport,omitempty"`
	DstPort int    `json:"dport,omitempty" yaml:"dport,omitempty"`
	Action  string `json:"action,omitempty" yaml:"action,omitempty"`
}

func (ru *ACLRule) Correct() {
	if ru.Action == "" {
		ru.Action = "drop"
	}
}
