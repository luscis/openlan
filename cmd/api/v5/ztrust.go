package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type ZTrust struct {
	Cmd
}

func (z ZTrust) Url(prefix, network, action string) string {
	return prefix + "/api/network/" + network + "/ztrust/" + action
}

func (z ZTrust) Enable(c *cli.Context) error {
	name := c.String("network")
	url := z.Url(c.String("url"), name, "enable")
	clt := z.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (z ZTrust) Disable(c *cli.Context) error {
	name := c.String("network")
	url := z.Url(c.String("url"), name, "disable")
	clt := z.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (z ZTrust) Commands(app *api.App) {
	name := api.GetUser(api.Token)
	user, network := api.SplitName(name)
	app.Command(&cli.Command{
		Name:  "ztrust",
		Usage: "Control Zero Trust",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "network", Value: network},
		},
		Subcommands: []*cli.Command{
			{
				Name:   "enable",
				Usage:  "Enable zTrust",
				Action: z.Enable,
			},
			{
				Name:   "disable",
				Usage:  "Disable zTrust",
				Action: z.Disable,
			},
			Guest{}.Commands(user),
			Knock{}.Commands(user),
		},
	})
}

type Guest struct {
	Cmd
}

func (u Guest) Url(prefix, network, name string) string {
	if name == "" {
		return prefix + "/api/network/" + network + "/guest"
	}
	return prefix + "/api/network/" + network + "/guest/" + name
}

func (u Guest) Add(c *cli.Context) error {
	guest := &schema.ZGuest{
		Address: c.String("address"),
		Name:    c.String("user"),
		Network: c.String("network"),
	}
	url := u.Url(c.String("url"), guest.Network, guest.Name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, guest, nil); err != nil {
		return err
	}
	return nil
}

func (u Guest) Remove(c *cli.Context) error {
	guest := &schema.ZGuest{
		Name:    c.String("user"),
		Network: c.String("network"),
		Address: c.String("address"),
	}
	url := u.Url(c.String("url"), guest.Network, guest.Name)
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

	url := u.Url(c.String("url"), network, "")
	clt := u.NewHttp(c.String("token"))

	var items []schema.ZGuest
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Guest) Commands(user string) *cli.Command {
	return &cli.Command{
		Name:  "guest",
		Usage: "zTrust Guest configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a zGuest",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Value: user},
					&cli.StringFlag{Name: "address"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing zGuest",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Value: user},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all zGuests",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	}
}

type Knock struct {
	Cmd
}

func (u Knock) Url(prefix, network, name string) string {
	return prefix + "/api/network/" + network + "/guest/" + name + "/knock"
}

func (u Knock) Add(c *cli.Context) error {
	socket := c.String("socket")
	knock := &schema.KnockRule{
		Protocol: c.String("protocol"),
		Age:      c.Int("age"),
		Name:     c.String("user"),
		Network:  c.String("network"),
	}
	knock.Dest, knock.Port = api.SplitSocket(socket)

	url := u.Url(c.String("url"), knock.Network, knock.Name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, knock, nil); err != nil {
		return err
	}
	return nil
}

func (u Knock) Remove(c *cli.Context) error {
	socket := c.String("socket")
	knock := &schema.KnockRule{
		Protocol: c.String("protocol"),
		Name:     c.String("user"),
		Network:  c.String("network"),
	}
	knock.Dest, knock.Port = api.SplitSocket(socket)

	url := u.Url(c.String("url"), knock.Network, knock.Name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, knock, nil); err != nil {
		return err
	}
	return nil
}

func (u Knock) Tmpl() string {
	return `# total {{ len . }}
{{ps -24 "username"}} {{ps -8 "protocol"}} {{ps -24 "socket"}} {{ps -4 "age"}} {{ps -24 "createAt"}}
{{- range . }}
{{p2 -24 "%s@%s" .Name .Network}} {{ps -8 .Protocol}} {{p2 -24 "%s:%s" .Dest .Port}} {{pi -4 .Age}} {{ut .CreateAt}}
{{- end }}
`
}

func (u Knock) List(c *cli.Context) error {
	network := c.String("network")
	user := c.String("user")

	url := u.Url(c.String("url"), network, user)
	clt := u.NewHttp(c.String("token"))

	var items []schema.KnockRule
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Knock) Commands(user string) *cli.Command {
	return &cli.Command{
		Name:  "knock",
		Usage: "Knock configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a knock",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Value: user},
					&cli.StringFlag{Name: "protocol", Required: true},
					&cli.StringFlag{Name: "socket", Required: true},
					&cli.IntFlag{Name: "age", Value: 60},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing knock",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Value: user},
					&cli.StringFlag{Name: "protocol", Required: true},
					&cli.StringFlag{Name: "socket", Required: true},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all knock",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Value: user},
				},
				Action: u.List,
			},
		},
	}
}
