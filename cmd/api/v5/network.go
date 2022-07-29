package v5

import (
	"fmt"
	"github.com/luscis/openlan/cmd/api"
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
	url += "?format=" + c.String("format")
	clt := u.NewHttp(c.String("token"))
	if data, err := clt.GetBody(url); err == nil {
		fmt.Println(string(data))
		return nil
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
