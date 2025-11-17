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
	return prefix + "/api/network/ceci/tcp"
}

func (u Ceci) Add(c *cli.Context) error {
	target := strings.Split(c.String("target"), ",")
	data := &schema.CeciTcp{
		Mode:   c.String("mode"),
		Listen: c.String("listen"),
		Target: target,
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u Ceci) Remove(c *cli.Context) error {
	data := &schema.CeciTcp{
		Listen: c.String("listen"),
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u Ceci) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "ceci",
		Usage: "Ceci TCP proxy",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a ceci TCP",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
					&cli.StringFlag{Name: "mode", Required: true},
					&cli.StringFlag{Name: "target"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a ceci TCP",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
				},
				Action: u.Remove,
			},
		},
	})
}
