package schema

type OnLine struct {
	HitTime    int64  `json:"aliveTime"`
	UpTime     int64  `json:"uptime"`
	EthType    uint16 `json:"ethType"`
	IpSource   string `json:"sourceAddr"`
	IpDest     string `json:"destAddr"`
	IpProto    string `json:"protocol"`
	PortSource uint16 `json:"sourcePort"`
	PortDest   uint16 `json:"destPort"`
}
