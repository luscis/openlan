package schema

type Qos struct {
	Name    string `json:"name"`
	Device  string `json:"device"`
	Ip      string `json:"ip"`
	InSpeed int64  `json:"inSpeed"`
}
