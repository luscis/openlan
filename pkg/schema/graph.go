package schema

type Category struct {
	Name string `json:"name"`
}

type Label struct {
	Show bool `json:"show"`
}

type GraphNode struct {
	Name       string `json:"name"`
	Value      int    `json:"value"`
	SymbolSize int    `json:"symbolSize"`
	Category   int    `json:"category"`
	Id         int    `json:"id"`
	Label      *Label `json:"label,omitempty"`
}

type GraphLink struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Weight int `json:"weight"`
}
