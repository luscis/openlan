package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Rate struct {
	Cmd
}

func (u Rate) Url(prefix, name string) string {
	return prefix + "/api/interface/" + name + "/rate"
}

func (u Rate) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "device"}} {{ps -8 "speed"}}
{{- range . }}
{{ps -16 .Device}} {{pi .Speed}}
{{- end }}
`
}

func (u Rate) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Rate
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Rate) Add(c *cli.Context) error {
	name := c.String("device")
	rate := &schema.Rate{
		Device: name,
		Speed:  c.Int("speed"),
	}
	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, rate, nil); err != nil {
		return err
	}
	return nil
}

func (u Rate) Remove(c *cli.Context) error {
	name := c.String("device")

	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u Rate) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "rate",
		Usage: "Rate Limit",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all rate limits",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add a rate limit",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
					&cli.StringFlag{Name: "speed", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a rate limit",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
				},
				Action: u.Remove,
			},
		},
	})
}
