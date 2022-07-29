package schema

type ACL struct {
	Name  string    `json:"name"`
	Rules []ACLRule `json:"rules"`
}

type ACLRule struct {
	Name    string `json:"name"`
	SrcIp   string `json:"src"`
	DstIp   string `json:"dst"`
	Proto   string `json:"proto"`
	SrcPort int    `json:"sport"`
	DstPort int    `json:"dport"`
	Action  string `json:"action"`
}
