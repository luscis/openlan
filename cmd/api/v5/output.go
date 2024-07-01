package v5

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Output struct {
	Cmd
}

func (o Output) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/output"
}

func (o Output) Add(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	output := &schema.Output{
		Network:  network,
		Remote:   c.String("remote"),
		Segment:  c.Int("segment"),
		Protocol: c.String("protocol"),
		DstPort:  c.Int("dstport"),
		Secret:   c.String("secret"),
		Crypt:    c.String("crypt"),
	}
	url := o.Url(c.String("url"), network)
	clt := o.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, output, nil); err != nil {
		return err
	}
	return nil
}

func (o Output) Remove(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	output := &schema.Output{
		Network: network,
		Device:  c.String("device"),
	}
	url := o.Url(c.String("url"), network)
	clt := o.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, output, nil); err != nil {
		return err
	}
	return nil
}

func (o Output) Save(c *cli.Context) error {
	network := c.String("name")
	url := o.Url(c.String("url"), network)

	clt := o.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (o Output) Tmpl() string {
	return `# total {{ len . }}
{{ps -24 "network"}} {{ps -15 "protocol"}} {{ps -15 "Remote"}} {{ps -15 "segment"}} {{ps -15 "device"}}
{{- range . }}
{{ps -24 .Network}} {{ps -15 .Protocol}} {{ps -15 .Remote}} {{pi -15 .Segment }} {{ps -15 .Device}}
{{- end }}
`
}

func (o Output) List(c *cli.Context) error {
	url := o.Url(c.String("url"), c.String("name"))
	clt := o.NewHttp(c.String("token"))
	var items []schema.Output
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return o.Out(items, c.String("format"), o.Tmpl())
}

func (o Output) Commands() *cli.Command {
	return &cli.Command{
		Name:  "output",
		Usage: "Output configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add an output for the network",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "remote", Required: true},
					&cli.IntFlag{Name: "segment"},
					&cli.StringFlag{Name: "protocol"},
					&cli.StringFlag{Name: "dstport"},
					&cli.StringFlag{Name: "secret"},
					&cli.StringFlag{Name: "crypt"},
				},
				Action: o.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an output from the network",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
				},
				Action: o.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all outputs of the network",
				Aliases: []string{"ls"},
				Action:  o.List,
			},
			{
				Name:    "save",
				Usage:   "Save all outputs",
				Aliases: []string{"sa"},
				Action:  o.Save,
			},
		},
	}
}
