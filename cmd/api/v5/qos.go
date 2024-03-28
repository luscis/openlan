package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Qos struct {
	Cmd
}

func (q Qos) Commands(app *api.App) {
	rule := QosRule{}
	app.Command(&cli.Command{
		Name:  "qos",
		Usage: "qos for client in network",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Subcommands: []*cli.Command{
			rule.Commands(),
		},
	})
}

type QosRule struct {
	Cmd
}

func (qr QosRule) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/qos"
}

func (qr QosRule) Add(c *cli.Context) error {
	name := c.String("name")
	url := qr.Url(c.String("url"), name)

	rule := &schema.Qos{
		Name:     c.String("clientname"),
		InSpeed:  c.Int64("inspeed"),
		OutSpeed: c.Int64("outspeed"),
	}

	clt := qr.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, rule, nil); err != nil {
		return err
	}

	return nil
}

func (qr QosRule) Remove(c *cli.Context) error {
	name := c.String("name")
	url := qr.Url(c.String("url"), name)

	rule := &schema.Qos{
		Name: c.String("clientname"),
	}

	clt := qr.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, rule, nil); err != nil {
		return err
	}

	return nil
}

func (qr QosRule) Tmpl() string {
	return `# total {{ len . }}
{{ps -15 "Name"}} {{ps -15 "Device"}} {{ps -15 "ip"}} {{ps -8 "InSpeed"}} {{ps -8 "OutSpeed"}}
{{- range . }}
{{ps -15 .Name}} {{ps -15 .Device}} {{ps -15 .Ip}} {{pi -8 .InSpeed}} {{pi -8 .OutSpeed}}
{{- end }}
`
}

func (qr QosRule) List(c *cli.Context) error {
	name := c.String("name")

	url := qr.Url(c.String("url"), name)
	clt := qr.NewHttp(c.String("token"))

	var items []schema.Qos
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return qr.Out(items, c.String("format"), qr.Tmpl())
}

func (qr QosRule) Save(c *cli.Context) error {
	name := c.String("name")
	url := qr.Url(c.String("url"), name)

	clt := qr.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (qr QosRule) Commands() *cli.Command {
	return &cli.Command{
		Name:  "rule",
		Usage: "Access Control Qos Rule",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new qos rule for client",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "clientname", Aliases: []string{"cn"}},
					&cli.StringFlag{Name: "inspeed", Aliases: []string{"is"}},
					&cli.StringFlag{Name: "outspeed", Aliases: []string{"os"}},
				},
				Action: qr.Add,
			},
			{
				Name:    "remove",
				Usage:   "remove a qos rule",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "clientname", Aliases: []string{"cn"}},
				},
				Action: qr.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all qos rules",
				Aliases: []string{"ls"},
				Action:  qr.List,
			},
			{
				Name:    "save",
				Usage:   "Save all qos rules",
				Aliases: []string{"sa"},
				Action:  qr.Save,
			},
		},
	}
}
