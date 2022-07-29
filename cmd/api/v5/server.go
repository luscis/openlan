package v5

import (
	"fmt"
	"github.com/luscis/openlan/cmd/api"
	"github.com/urfave/cli/v2"
)

type Server struct {
	Cmd
}

func (u Server) Url(prefix, name string) string {
	return prefix + "/api/server"
}

func (u Server) List(c *cli.Context) error {
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

func (u Server) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "server",
		Aliases: []string{"sr"},
		Usage:   "Socket server status",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display server status",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
