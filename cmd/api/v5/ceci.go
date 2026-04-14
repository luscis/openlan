package v5

import (
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Ceci struct {
	Cmd
}

func (u Ceci) Url(prefix string) string {
	return prefix + "/api/network/ceci"
}

func (u Ceci) List(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	var data schema.Network
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return u.Out(data, "yaml", "")
}

func (u Ceci) Save(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u Ceci) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "ceci",
		Usage:  "Special Ceci proxy network",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List a Ceci proxy",
				Action: u.List,
			},
			{
				Name:    "save",
				Usage:   "Save Ceci network",
				Aliases: []string{"sa"},
				Action:  u.Save,
			},
			CeciProxy{}.Commands(app),
		},
	})
}

type CeciProxy struct {
	Cmd
}

func (u CeciProxy) Url(prefix string) string {
	return prefix + "/api/network/ceci/proxy"
}

func (u CeciProxy) Add(c *cli.Context) error {
	target := strings.Split(c.String("target"), ",")
	data := &schema.CeciProxy{
		Mode:    c.String("mode"),
		Listen:  c.String("listen"),
		Network: c.String("network"),
		Target:  target,
	}
	if cert := c.String("cert"); cert != "" || c.String("key") != "" || c.String("root-ca") != "" || c.Bool("insecure") {
		data.Cert = &schema.Cert{
			CrtFile:  cert,
			KeyFile:  c.String("key"),
			CaFile:   c.String("root-ca"),
			Insecure: c.Bool("insecure"),
		}
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u CeciProxy) Remove(c *cli.Context) error {
	data := &schema.CeciProxy{
		Listen: c.String("listen"),
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u CeciProxy) Commands(app *api.App) *cli.Command {
	return &cli.Command{
		Name:  "ceci",
		Usage: "Special Ceci proxy",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a Ceci Proxy",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
					&cli.StringFlag{Name: "mode", Required: true},
					&cli.StringFlag{Name: "network"},
					&cli.StringFlag{Name: "target"},
					&cli.StringFlag{Name: "cert"},
					&cli.StringFlag{Name: "key"},
					&cli.StringFlag{Name: "root-ca"},
					&cli.BoolFlag{Name: "insecure"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a Ceci Proxy",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
				},
				Action: u.Remove,
			},
		},
	}
}
