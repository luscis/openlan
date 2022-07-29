package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Device struct {
	Cmd
}

func (u Device) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/device"
	} else {
		return prefix + "/api/device/" + name
	}
}

func (u Device) Tmpl() string {
	return `# total {{ len . }}
{{ps -13 "name"}} {{ps -13 "mtu"}} {{ps -16 "mac"}} {{ps -6 "provider"}}
{{- range . }}
{{ps -13 .Name}} {{pi -13 .Mtu}} {{ps -16 .Mac}} {{ps -6 .Provider}}
{{- end }}
`
}

func (u Device) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Device
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Device) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "device",
		Aliases: []string{"dev"},
		Usage:   "linux network device",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all devices",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
