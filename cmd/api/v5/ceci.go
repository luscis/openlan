package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Ceci struct {
	Cmd
}

func (u Ceci) Url(prefix, name string) string {
	return prefix + "/api/interface/" + name + "/ceci"
}

func (u Ceci) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "Name"}} {{ps -8 "Configure"}}
{{- range . }}
{{ps -16 .Name}} {{pi .Configure}}
{{- end }}
`
}

func (u Ceci) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Ceci
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Ceci) Add(c *cli.Context) error {
	name := c.String("name")
	rate := &schema.Ceci{
		Name: name,
	}
	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, rate, nil); err != nil {
		return err
	}
	return nil
}

func (u Ceci) Remove(c *cli.Context) error {
	name := c.String("name")

	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u Ceci) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "ceci",
		Usage: "Ceci proxy",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all ceci proxy",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add a ceci proxy",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true},
					&cli.StringFlag{Name: "file", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a ceci proxy",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true},
				},
				Action: u.Remove,
			},
		},
	})
}
