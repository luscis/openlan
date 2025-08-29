package schema

type IPSecTunnel struct {
	Left      string `json:"local"`
	LeftId    string `json:"localid,omitempty"`
	LeftPort  int    `json:"localport,omitempty"`
	Right     string `json:"remote"`
	RightId   string `json:"remoteid,omitempty"`
	RightPort int    `json:"remoteport,omitempty"`
	Transport string `json:"protocol"`
	Secret    string `json:"secret"`
	State     string `json:"state"`
}
