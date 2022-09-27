package config

type ACL struct {
	File  string     `json:"file"`
	Name  string     `json:"name"`
	Rules []*ACLRule `json:"rules"`
}

type ACLRule struct {
	Name    string `json:"name,omitempty"`
	SrcIp   string `json:"source,omitempty"`
	DstIp   string `json:"destination,omitempty"`
	Proto   string `json:"protocol,omitempty"`
	SrcPort string `json:"sourcePort,omitempty"`
	DstPort string `json:"destPort,omitempty"`
	Action  string `json:"action,omitempty"`
}

func (ru *ACLRule) Correct() {
}
