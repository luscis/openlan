package schema

type Rate struct {
	Device string `json:"device"`
	Speed  int    `json:"speed"` // Mbit
}
