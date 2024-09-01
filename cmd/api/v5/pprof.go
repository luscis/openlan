package v5

import (
	"fmt"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type PProf struct {
	Cmd
}

func (u PProf) Url(prefix, name string) string {
	return prefix + "/api/pprof"
}

func (u PProf) Add(c *cli.Context) error {
	pp := schema.PProf{
		Listen: c.String("listen"),
	}
	if pp.Listen == "" {
		return libol.NewErr("listen value is empty")
	}
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, pp, nil); err != nil {
		return err
	}
	return nil
}

func (u PProf) Del(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u PProf) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var pp schema.PProf
	if err := clt.GetJSON(url, &pp); err != nil {
		return err
	}
	fmt.Println(pp.Listen)
	return nil
}

func (u PProf) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "pprof",
		Usage: "Debug pprof tool",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Show configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "enable",
				Usage: "Enable pprof tool",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Value: "127.0.0.1:6060"},
				},
				Action: u.Add,
			},
		},
	})
}
