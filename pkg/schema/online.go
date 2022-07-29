package schema

type OnLine struct {
	HitTime    int64  `json:"hittime"`
	UpTime     int64  `json:"uptime"`
	EthType    uint16 `json:"ethType"`
	IpSource   string `json:"ipSource"`
	IpDest     string `json:"ipDestination"`
	IpProto    string `json:"ipProtocol"`
	PortSource uint16 `json:"portSource"`
	PortDest   uint16 `json:"portDestination"`
}
