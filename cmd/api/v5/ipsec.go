package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Esp struct {
	Cmd
}

func (u Esp) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/esp"
	} else {
		return prefix + "/api/esp/" + name
	}
}

func (u Esp) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "name"}} {{ps -16 "address"}} 
{{- range . }}
{{ps -16 .Name}} {{ps -16 .Address}}
{{- end }}
`
}

func (u Esp) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Esp
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Esp) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "esp",
		Aliases: []string{"esp"},
		Usage:   "IPSec ESP configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all esp",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
