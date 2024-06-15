package schema

type IPSecTunnel struct {
	Left      string `json:"local"`
	LeftId    string `json:"localid"`
	LeftPort  string `json:"localport"`
	Right     string `json:"remote"`
	RightId   string `json:"remoteid"`
	RightPort string `json:"remoteport"`
	Transport string `json:"transport"`
	Secret    string `json:"secret"`
}
