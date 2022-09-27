package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Network struct {
	Cmd
}

func (u Network) Url(prefix, name string) string {
	return prefix + "/api/network"
}

func (u Network) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Network
	if err := clt.GetJSON(url, &items); err == nil {
		return u.Out(items, c.String("format"), "")
	} else {
		return err
	}
}

func (u Network) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "network",
		Aliases: []string{"net"},
		Usage:   "Logical network",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all network",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
