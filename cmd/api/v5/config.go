package v5

import (
	"fmt"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Cmd
}

func (u Config) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/config"
	}
	return prefix + "/api/config/" + name
}

func (u Config) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	cfg := &config.Switch{}
	if err := clt.GetJSON(url, cfg); err == nil {
		name := c.String("network")
		format := c.String("format")
		cfg.RemarshalNetworks(format)
		if len(name) > 0 {
			obj := cfg.GetNetwork(name)
			return u.Out(obj, format, "")
		} else {
			return u.Out(cfg, format, "")
		}
	} else {
		return err
	}
}

func (u Config) Reload(c *cli.Context) error {
	url := u.Url(c.String("url"), "reload")
	clt := u.NewHttp(c.String("token"))
	data := &schema.Message{}
	if err := clt.PutJSON(url, nil, data); err == nil {
		fmt.Println(data.Message)
		return nil
	} else {
		return err
	}
}

func (u Config) Save(c *cli.Context) error {
	url := u.Url(c.String("url"), "save")
	clt := u.NewHttp(c.String("token"))
	data := &schema.Message{}
	if err := clt.PutJSON(url, nil, data); err == nil {
		fmt.Println(data.Message)
		return nil
	} else {
		return err
	}
}

func (u Config) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "config",
		Usage:  "Switch configuration",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all configuration",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Value: ""},
				},
				Action: u.List,
			},
			{
				Name:    "reload",
				Usage:   "Reload configuration",
				Aliases: []string{"re"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dir", Value: "/etc/openlan"},
				},
				Action: u.Reload,
			},
			{
				Name:    "save",
				Usage:   "Save configuration",
				Aliases: []string{"sa"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dir", Value: "/etc/openlan"},
				},
				Action: u.Save,
			},
		},
	})
}
