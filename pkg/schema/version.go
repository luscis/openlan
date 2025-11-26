package schema

import "github.com/luscis/openlan/pkg/libol"

type Version struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
	Expire  string `json:"expire"`
}

func NewVersionSchema() Version {
	return Version{
		Version: libol.Version,
		Date:    libol.Date,
		Commit:  libol.Commit,
	}
}

type VersionCert struct {
	Ca         string `json:"ca,omitempty"`
	Cert       string `json:"crt,omitempty"`
	Key        string `json:"key,omitempty"`
	CertExpire string `json:"crtExpire,omitempty"`
	CaExpire   string `json:"caExpire,omitempty"`
}
