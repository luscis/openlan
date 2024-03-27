package config

import "github.com/luscis/openlan/pkg/libol"

type ACL struct {
	File  string     `json:"file"`
	Name  string     `json:"name"`
	Rules []*ACLRule `json:"rules"`
}

func (ru *ACL) Save() {
	if err := libol.MarshalSave(ru, ru.File, true); err != nil {
		libol.Error("Switch.Save.Acl %s %s", ru.Name, err)
	}
}

type ACLRule struct {
	Name    string `json:"name,omitempty"`
	SrcIp   string `json:"source,omitempty"`
	DstIp   string `json:"destination,omitempty"`
	Proto   string `json:"protocol,omitempty"`
	SrcPort int    `json:"sport,omitempty"`
	DstPort int    `json:"dport,omitempty"`
	Action  string `json:"action,omitempty"`
}

func (ru *ACLRule) Correct() {
}
