package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

// openlan router tunnel add --remote 1.1.1.2 --address 192.168.1.1 --protocol gre
// openlan router tunnel add --remote 1.1.1.2 --address 192.168.1.1 --protocol ipip
// openlan router tunnel remove --remote 1.1.1.2 --address 192.168.1.1

type Router struct {
	Cmd
}

func (b Router) Url(prefix string) string {
	return prefix + "/api/network/router"
}

func (b Router) List(c *cli.Context) error {
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	var data schema.Network
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return b.Out(data, "yaml", "")
}

func (b Router) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "router",
		Usage: "Router",
		Subcommands: []*cli.Command{
			{
				Name:    "ls",
				Usage:   "Display router network",
				Aliases: []string{"ls"},
				Action:  b.List,
			},
			RouterTunnel{}.Commands(),
		},
	})
}

type RouterTunnel struct {
	Cmd
}

func (s RouterTunnel) Url(prefix string) string {
	return prefix + "/api/network/router/tunnel"
}

func (s RouterTunnel) Add(c *cli.Context) error {
	data := &schema.RouterTunnel{
		Remote:   c.String("remote"),
		Address:  c.String("address"),
		Protocol: c.String("protocol"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterTunnel) Remove(c *cli.Context) error {
	data := &schema.RouterTunnel{
		Remote:   c.String("remote"),
		Protocol: c.String("protocol"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterTunnel) Commands() *cli.Command {
	return &cli.Command{
		Name:  "tunnel",
		Usage: "Router tunnels",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add router tunnel",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "address", Required: true},
					&cli.StringFlag{Name: "remote", Required: true},
					&cli.StringFlag{Name: "protocol", Value: "gre"},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove router tunnel",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "remote", Required: true},
					&cli.StringFlag{Name: "protocol", Value: "gre"},
				},
				Action: s.Remove,
			},
		},
	}
}
