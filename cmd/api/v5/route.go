package v5

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Route struct {
	Cmd
}

func (r Route) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/route"
}

func (r Route) Add(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	pr := &schema.PrefixRoute{
		Prefix:  c.String("prefix"),
		NextHop: c.String("nexthop"),
		FindHop: c.String("findhop"),
		Metric:  c.Int("metric"),
		Mode:    c.String("mode"),
	}
	url := r.Url(c.String("url"), network)
	clt := r.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, pr, nil); err != nil {
		return err
	}
	return nil
}

func (r Route) Remove(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	pr := &schema.PrefixRoute{
		Prefix: c.String("prefix"),
	}
	url := r.Url(c.String("url"), network)
	clt := r.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, pr, nil); err != nil {
		return err
	}
	return nil
}

func (r Route) Save(c *cli.Context) error {
	network := c.String("name")
	url := r.Url(c.String("url"), network)

	clt := r.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (r Route) Tmpl() string {
	return `# total {{ len . }}
{{ps -25 "prefix"}} {{ps -25 "nexthop"}} {{ps -8 "metric"}} {{ps -8 "mode"}}
{{- range . }}
{{ps -25 .Prefix}} {{ if .FindHop }}{{ps -25 .FindHop}}{{ else }}{{ps -25 .NextHop}}{{ end }} {{pi -8 .Metric }} {{ps -8 .Mode}}
{{- end }}
`
}

func (r Route) List(c *cli.Context) error {
	url := r.Url(c.String("url"), c.String("name"))
	clt := r.NewHttp(c.String("token"))
	var items []schema.PrefixRoute
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return r.Out(items, c.String("format"), r.Tmpl())
}

func (r Route) Commands() *cli.Command {
	return &cli.Command{
		Name:  "route",
		Usage: "Prefix route",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a route for the network",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "prefix", Required: true},
					&cli.StringFlag{Name: "nexthop"},
					&cli.StringFlag{Name: "findhop"},
					&cli.IntFlag{Name: "metric"},
					&cli.StringFlag{Name: "mode"},
				},
				Action: r.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a route from the network",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "prefix", Required: true},
				},
				Action: r.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all routes of the network",
				Aliases: []string{"ls"},
				Action:  r.List,
			},
			{
				Name:    "save",
				Usage:   "Save all routes",
				Aliases: []string{"sa"},
				Action:  r.Save,
			},
		},
	}
}
