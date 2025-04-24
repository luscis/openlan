package config

type FlowRule struct {
	Table    string `json:"table,omitempty" yaml:"table,omitempty"`
	Chain    string `json:"chain,omitempty" yaml:"chain,omitempty"`
	Input    string `json:"input,omitempty" yaml:"input,omitempty"`
	Source   string `json:"source,omitempty" yaml:"source,omitempty"`
	ToSource string `json:"toSource,omitempty" yaml:"toSource,omitempty"`
	Dest     string `json:"destination,omitempty" yaml:"destination,omitempty"`
	ToDest   string `json:"toDestination" yaml:"toDestination"`
	Output   string `json:"output,omitempty" yaml:"output,omitempty"`
	Comment  string `json:"comment,omitempty" yaml:"comment,omitempty"`
	Proto    string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Match    string `json:"match,omitempty" yaml:"match,omitempty"`
	DstPort  string `json:"destPort,omitempty" yaml:"destPort,omitempty"`
	SrcPort  string `json:"sourcePort,omitempty" yaml:"sourcePort,omitempty"`
	CtState  string `json:"ctState,omitempty" yaml:"ctState,omitempty"`
	Jump     string `json:"jump,omitempty" yaml:"jump,omitempty"` // SNAT/RETURN/MASQUERADE
}
