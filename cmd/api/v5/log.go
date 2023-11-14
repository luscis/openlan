package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Log struct {
	Cmd
}

func (v Log) Url(prefix, name string) string {
	return prefix + "/api/log"
}

func (v Log) Tmpl() string {
	return `File :  {{ .File }}
Level:  {{ .Level}}
`
}

func (v Log) List(c *cli.Context) error {
	url := v.Url(c.String("url"), "")
	clt := v.NewHttp(c.String("token"))
	var item schema.Log
	if err := clt.GetJSON(url, &item); err != nil {
		return err
	}
	return v.Out(item, c.String("format"), v.Tmpl())
}

func (v Log) Add(c *cli.Context) error {
	url := v.Url(c.String("url"), "")
	log := &schema.Log{
		Level: c.Int("level"),
	}
	clt := v.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, log, nil); err != nil {
		return err
	}
	return nil
}

func (v Log) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "log",
		Aliases: []string{"v"},
		Usage:   "show log information",
		Action:  v.List,
		Subcommands: []*cli.Command{
			{
				Name:  "set",
				Usage: "set log level",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "level"},
				},
				Action: v.Add,
			},
		},
	})
}
