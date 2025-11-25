package v5

import (
	"os"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Version struct {
	Cmd
}

func (v Version) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/version"
	} else {
		return prefix + "/api/version/" + name
	}
}

func (v Version) Tmpl() string {
	return `Version  :  {{ .Version }}
Build at :  {{ .Date}}
Expire at:  {{ .Expire }}
`
}

func (v Version) List(c *cli.Context) error {
	url := v.Url(c.String("url"), "")
	clt := v.NewHttp(c.String("token"))
	var item schema.Version
	if err := clt.GetJSON(url, &item); err != nil {
		return err
	}
	return v.Out(item, c.String("format"), v.Tmpl())
}

func (v Version) Update(c *cli.Context) error {
	data := &schema.VersionCert{}
	if c.String("cert") != "" {
		value, err := os.ReadFile(c.String("cert"))
		if err != nil {
			return err
		}
		data.Cert = string(value)
	}
	if c.String("ca") != "" {
		value, err := os.ReadFile(c.String("ca"))
		if err != nil {
			return err
		}
		data.Ca = string(value)
	}
	if c.String("key") != "" {
		value, err := os.ReadFile(c.String("key"))
		if err != nil {
			return err
		}
		data.Ca = string(value)
	}

	url := v.Url(c.String("url"), "cert")
	clt := v.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (v Version) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "version",
		Usage:  "show version information",
		Action: v.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all network",
				Aliases: []string{"ls"},
				Action:  v.List,
			},
			{
				Name:  "update",
				Usage: "update cert from file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "cert"},
					&cli.StringFlag{Name: "key"},
					&cli.StringFlag{Name: "ca"},
				},
				Action: v.Update,
			},
		},
	})
}
