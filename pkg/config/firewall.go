package config

type FlowRule struct {
	Table    string `json:"table,omitempty"`
	Chain    string `json:"chain,omitempty"`
	Input    string `json:"input,omitempty"`
	Source   string `json:"source,omitempty"`
	ToSource string `json:"toSource,omitempty"`
	Dest     string `json:"destination,omitempty"`
	ToDest   string `json:"toDestination"`
	Output   string `json:"output,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Proto    string `json:"protocol,omitempty"`
	Match    string `json:"match,omitempty"`
	DstPort  string `json:"destPort,omitempty"`
	SrcPort  string `json:"sourcePort,omitempty"`
	CtState  string `json:"ctState,omitempty"`
	Jump     string `json:"jump,omitempty"` // SNAT/RETURN/MASQUERADE
}
