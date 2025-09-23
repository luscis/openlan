package schema

import "github.com/luscis/openlan/pkg/libol"

type Version struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Expire  string `json:"expire"`
}

func NewVersionSchema() Version {
	return Version{
		Version: libol.Version,
		Date:    libol.Date,
	}
}
