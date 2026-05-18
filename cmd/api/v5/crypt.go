package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Crypt struct {
	Cmd
}

func (v Crypt) Url(prefix string) string {
	return prefix + "/api/version/crypt"
}

func (v Crypt) List(c *cli.Context) error {
	url := v.Url(c.String("url"))
	clt := v.NewHttp(c.String("token"))
	var item schema.SwitchCrypt
	if err := clt.GetJSON(url, &item); err != nil {
		return err
	}
	return v.Out(item, "yaml", "")
}

func (v Crypt) Update(c *cli.Context) error {
	data := &schema.SwitchCrypt{
		Algorithm: c.String("algorithm"),
		Secret:    c.String("secret"),
	}
	url := v.Url(c.String("url"))
	clt := v.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (v Crypt) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "crypt",
		Usage:  "Switch crypt configuration",
		Action: v.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "Display crypt details",
				Action:  v.List,
			},
			{
				Name:  "update",
				Usage: "Update crypt algorithm and secret",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "algorithm"},
					&cli.StringFlag{Name: "secret"},
				},
				Action: v.Update,
			},
		},
	})
}
