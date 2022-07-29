package schema

type VxLAN struct {
	Name    string        `json:"name"`
	Bridge  string        `json:"bridge"`
	Members []VxLANMember `json:"members"`
}

type VxLANMember struct {
	Vni    int    `json:"vni"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
}
