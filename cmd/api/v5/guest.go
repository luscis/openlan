package v5

import (
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Guest struct {
	Cmd
}

func (u Guest) Url(prefix, name string) string {
	name, network := api.SplitName(name)
	if network == "" {
		return prefix + "/api/network/" + name + "/guest"
	}
	return prefix + "/api/network/" + network + "/guest/" + name
}

func (u Guest) Add(c *cli.Context) error {
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

func (u Guest) Remove(c *cli.Context) error {
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

func (u Guest) Tmpl() string {
	return `# total {{ len . }}
{{ps -24 "username"}} {{ps -24 "address"}}
{{- range . }}
{{p2 -24 "%s@%s" .Name .Network}} {{ps -24 .Address}}
{{- end }}
`
}

func (u Guest) List(c *cli.Context) error {
	network := c.String("network")

	url := u.Url(c.String("url"), network)
	clt := u.NewHttp(c.String("token"))

	var items []schema.ZGuest
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Guest) Commands(app *api.App) {
	name := api.GetUser(api.Token)
	app.Command(&cli.Command{
		Name:    "guest",
		Aliases: []string{"gu"},
		Usage:   "ZTrust Guest configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a zGuest",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Value: name},
					&cli.StringFlag{Name: "address"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing zGuest",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Value: name},
					&cli.StringFlag{Name: "address"},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all zGuests",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Value: name},
				},
				Action: u.List,
			},
		},
	})
}
