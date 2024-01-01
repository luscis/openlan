package v5

import (
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type ZGuest struct {
	Cmd
}

func (u ZGuest) Url(prefix, name string) string {
	name, network := api.SplitName(name)
	if name == "" {
		return prefix + "/api/ztrust/" + network + "/guest"
	} else {
		return prefix + "/api/ztrust/" + network + "/guest/" + name
	}
}

func (u ZGuest) Add(c *cli.Context) error {
	username := c.String("name")
	if !strings.Contains(username, "@") {
		return libol.NewErr("invalid username")
	}
	guest := &schema.ZGuest{
		Name:    username,
		Address: c.String("address"),
	}
	guest.Name, guest.Network = api.SplitName(username)
	url := u.Url(c.String("url"), username)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, guest, nil); err != nil {
		return err
	}
	return nil
}

func (u ZGuest) Remove(c *cli.Context) error {
	username := c.String("name")
	if !strings.Contains(username, "@") {
		return libol.NewErr("invalid username")
	}
	guest := &schema.ZGuest{
		Name:    username,
		Address: c.String("address"),
	}
	guest.Name, guest.Network = api.SplitName(username)
	url := u.Url(c.String("url"), username)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, guest, nil); err != nil {
		return err
	}
	return nil
}

func (u ZGuest) Tmpl() string {
	return `# total {{ len . }}
{{ps -24 "username"}} {{ps -24 "address"}}
{{- range . }}
{{p2 -24 "%s@%s" .Name .Network}} {{ps -24 .Address}}
{{- end }}
`
}

func (u ZGuest) List(c *cli.Context) error {
	return nil
}

func (u ZGuest) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "zguest",
		Aliases: []string{"zg"},
		Usage:   "zGuest configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a zGuest",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "address"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing zGuest",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "address"},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all zGuests",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network"},
				},
				Action: u.List,
			},
		},
	})
}
