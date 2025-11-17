package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Version struct {
	Cmd
}

func (v Version) Url(prefix, name string) string {
	return prefix + "/api/version"
}

func (v Version) Tmpl() string {
	return `Version  :  {{ .Version }}
Build at :  {{ .Date}}
Expire at:  {{ .Expire }}
`
}

func (v Version) List(c *cli.Context) error {
	url := v.Url(c.String("url"), "")
	clt := v.NewHttp(c.String("token"))
	var item schema.Version
	if err := clt.GetJSON(url, &item); err != nil {
		return err
	}
	return v.Out(item, c.String("format"), v.Tmpl())
}

func (v Version) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "version",
		Usage:  "show version information",
		Action: v.List,
	})
}
