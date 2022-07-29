package schema

type Worker struct {
	Uptime   int64  `json:"uptime"`
	UUID     string `json:"uuid"`
	Alias    string `json:"alias"`
	Protocol string `json:"protocol"`
}
