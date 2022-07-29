package schema

type Switch struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Address string `json:"address"`
}
