package config

type Qos struct {
	Config map[string]QosLimit `json:"qos,omitempty"`
}

type QosLimit struct {
	InSpeed  int64 `json:"inSpeed,omitempty"`
	OutSpeed int64 `json:"outSpeed,omitempty"`
}
