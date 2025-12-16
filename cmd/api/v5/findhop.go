package v5

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type FindHop struct {
	Cmd
}

func (r FindHop) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/findhop"
}

func (r FindHop) Add(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	pr := &schema.FindHop{
		Name:    c.String("findhop"),
		Mode:    c.String("mode"),
		NextHop: c.String("nexthop"),
		Check:   c.String("check"),
	}
	url := r.Url(c.String("url"), network)
	clt := r.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, pr, nil); err != nil {
		return err
	}
	return nil
}

func (r FindHop) Remove(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	pr := &schema.FindHop{
		Name: c.String("findhop"),
	}
	url := r.Url(c.String("url"), network)
	clt := r.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, pr, nil); err != nil {
		return err
	}
	return nil
}

func (r FindHop) Save(c *cli.Context) error {
	network := c.String("name")
	url := r.Url(c.String("url"), network)

	clt := r.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (r FindHop) Tmpl() string {
	return `# total {{ len . }}
{{ps -25 "name"}} {{ps -8 "checker"}} {{ps -16 "mode"}} {{ps -25 "nexthop"}} {{ps -25 "available"}}
{{- range . }}
{{ps -25 .Name}} {{ps -8 .Check }} {{ps -16 .Mode}} {{ps -25 .NextHop}} {{ps -25 .Available}} 
{{- end }}
`
}

func (r FindHop) List(c *cli.Context) error {
	url := r.Url(c.String("url"), c.String("name"))
	clt := r.NewHttp(c.String("token"))
	var items []schema.FindHop
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return r.Out(items, c.String("format"), r.Tmpl())
}

func (r FindHop) Commands() *cli.Command {
	return &cli.Command{
		Name:   "findhop",
		Usage:  "FindHop configuration",
		Action: r.List,
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a findhop for the network",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "findhop", Required: true},
					&cli.StringFlag{Name: "nexthop"},
					&cli.StringFlag{Name: "mode", Value: "active-backup"},
					&cli.StringFlag{Name: "check"},
				},
				Action: r.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a findhop from the network",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "findhop", Required: true},
				},
				Action: r.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all findhop of the network",
				Aliases: []string{"ls"},
				Action:  r.List,
			},
			{
				Name:    "save",
				Usage:   "Save all findhop",
				Aliases: []string{"sa"},
				Action:  r.Save,
			},
		},
	}
}
