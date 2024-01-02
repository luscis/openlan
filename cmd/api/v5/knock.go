package v5

import (
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Knock struct {
	Cmd
}

func (u Knock) Url(prefix, name string) string {
	name, network := api.SplitName(name)
	return prefix + "/api/network/" + network + "/guest/" + name + "/knock"
}

func (u Knock) Add(c *cli.Context) error {
	username := c.String("name")
	if !strings.Contains(username, "@") {
		return libol.NewErr("invalid username")
	}
	socket := c.String("socket")
	knock := &schema.KnockRule{
		Protocol: c.String("protocol"),
	}
	knock.Name, knock.Network = api.SplitName(username)
	knock.Dest, knock.Port = api.SplitSocket(socket)

	url := u.Url(c.String("url"), username)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, knock, nil); err != nil {
		return err
	}
	return nil
}

func (u Knock) Remove(c *cli.Context) error {
	username := c.String("name")
	if !strings.Contains(username, "@") {
		return libol.NewErr("invalid username")
	}
	socket := c.String("socket")
	knock := &schema.KnockRule{
		Protocol: c.String("protocol"),
	}
	knock.Name, knock.Network = api.SplitName(username)
	knock.Dest, knock.Port = api.SplitSocket(socket)

	url := u.Url(c.String("url"), username)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, knock, nil); err != nil {
		return err
	}
	return nil
}

func (u Knock) Tmpl() string {
	return `# total {{ len . }}
{{ps -24 "username"}} {{ps -8 "protocol"}} {{ps -24 "socket"}} {{ps -24 "createAt"}}
{{- range . }}
{{p2 -24 "%s@%s" .Name .Network}} {{ps -8 .Protocol}} {{p2 -24 "%s:%s" .Dest .Port}} {{ut .CreateAt}}
{{- end }}
`
}

func (u Knock) List(c *cli.Context) error {
	name := c.String("name")

	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))

	var items []schema.KnockRule
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Knock) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "knock",
		Aliases: []string{"kn"},
		Usage:   "Knock configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a knock",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "protocol"},
					&cli.StringFlag{Name: "socket"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing knock",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "protocol"},
					&cli.StringFlag{Name: "socket"},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all knock",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
				Action: u.List,
			},
		},
	})
}
