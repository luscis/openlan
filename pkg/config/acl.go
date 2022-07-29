package config

type ACL struct {
	File  string     `json:"file"`
	Name  string     `json:"name"`
	Rules []*ACLRule `json:"rules"`
}

type ACLRule struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	SrcIp   string `json:"src,omitempty" yaml:"source,omitempty"`
	DstIp   string `json:"dst,omitempty" yaml:"destination,omitempty"`
	Proto   string `json:"proto,omitempty" yaml:"protocol,omitempty"`
	SrcPort string `json:"sport,omitempty" yaml:"destPort,omitempty"`
	DstPort string `json:"dport,omitempty" yaml:"sourcePort,omitempty"`
	Action  string `json:"action,omitempty" yaml:"action,omitempty"`
}

func (ru *ACLRule) Correct() {
}
