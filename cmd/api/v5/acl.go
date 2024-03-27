package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type ACL struct {
	Cmd
}

func (u ACL) Commands(app *api.App) {
	rule := ACLRule{}
	app.Command(&cli.Command{
		Name:  "acl",
		Usage: "Access control list",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Subcommands: []*cli.Command{
			rule.Commands(),
		},
	})
}

type ACLRule struct {
	Cmd
}

func (u ACLRule) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/acl"
}

func (u ACLRule) Add(c *cli.Context) error {
	name := c.String("name")
	url := u.Url(c.String("url"), name)

	rule := &schema.ACLRule{
		Proto:   c.String("protocol"),
		SrcIp:   c.String("source"),
		DstIp:   c.String("destination"),
		SrcPort: c.Int("sport"),
		DstPort: c.Int("dport"),
		Action:  "DROP",
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, rule, nil); err != nil {
		return err
	}

	return nil
}

func (u ACLRule) Remove(c *cli.Context) error {
	name := c.String("name")
	url := u.Url(c.String("url"), name)

	rule := &schema.ACLRule{
		Proto:   c.String("protocol"),
		SrcIp:   c.String("source"),
		DstIp:   c.String("destination"),
		SrcPort: c.Int("sport"),
		DstPort: c.Int("dport"),
		Action:  "DROP",
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, rule, nil); err != nil {
		return err
	}

	return nil
}

func (u ACLRule) Tmpl() string {
	return `# total {{ len . }}
{{ps -15 "source"}} {{ps -15 "destination"}} {{ps -4 "protocol"}} {{ps -4 "dport"}} {{ps -4 "sport"}}
{{- range . }}
{{ps -15 .SrcIp}} {{ps -15 .DstIp}} {{ps -4 .Proto}} {{pi -4 .DstPort}} {{pi -4 .SrcPort}}
{{- end }}
`
}

func (u ACLRule) List(c *cli.Context) error {
	name := c.String("name")

	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))

	var items []schema.ACLRule
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}

	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u ACLRule) Commands() *cli.Command {
	return &cli.Command{
		Name:  "rule",
		Usage: "Access control list rule",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new acl rule",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "source", Aliases: []string{"s"}},
					&cli.StringFlag{Name: "destination", Aliases: []string{"d"}},
					&cli.StringFlag{Name: "protocol", Aliases: []string{"p"}},
					&cli.IntFlag{Name: "sport", Aliases: []string{"dp"}},
					&cli.IntFlag{Name: "dport", Aliases: []string{"sp"}},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "remove a new acl rule",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "source", Aliases: []string{"s"}},
					&cli.StringFlag{Name: "destination", Aliases: []string{"d"}},
					&cli.StringFlag{Name: "protocol", Aliases: []string{"p"}},
					&cli.IntFlag{Name: "sport", Aliases: []string{"dp"}},
					&cli.IntFlag{Name: "dport", Aliases: []string{"sp"}},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all acl rules",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	}
}
