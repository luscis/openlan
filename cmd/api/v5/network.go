package v5

import (
	"github.com/luscis/openlan/cmd/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Network struct {
	Cmd
}

func (u Network) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/network"
	} else {
		return prefix + "/api/network/" + name
	}
}

func (u Network) List(c *cli.Context) error {
	name := c.String("name")
	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))

	if name == "" {
		var items []schema.Network
		if err := clt.GetJSON(url, &items); err == nil {
			return u.Out(items, c.String("format"), "")
		} else {
			return err
		}
	} else {
		var items schema.Network
		if err := clt.GetJSON(url, &items); err == nil {
			return u.Out(items, c.String("format"), "")
		} else {
			return err
		}
	}
}

func (u Network) Add(c *cli.Context) error {
	file := c.String("file")
	name := c.String("name")
	config := &co.Network{}
	network := &schema.Network{Config: config}

	if file == "" && name == "" {
		return libol.NewErr("invalid network")
	}
	if file != "" {
		if err := libol.UnmarshalLoad(&network.Config, file); err != nil {
			return err
		}
	}
	if name != "" {
		config.Name = name
	}
	if c.String("address") != "" {
		if config.Bridge == nil {
			config.Bridge = &co.Bridge{}
		}
		config.Bridge.Address = c.String("address")
	}
	if c.String("provider") != "" {
		config.Provider = c.String("provider")
	}
	if c.String("namespace") != "" {
		config.Namespace = c.String("namespace")
	}
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, network, nil); err != nil {
		return err
	}
	return nil
}

func (u Network) Remove(c *cli.Context) error {
	network := c.String("name")
	if network == "" {
		return libol.NewErr("invalid network")
	}
	url := u.Url(c.String("url"), network)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u Network) Save(c *cli.Context) error {
	name := c.String("name")
	network := &schema.Network{
		Name: name,
	}
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, network, nil); err != nil {
		return err
	}

	return nil
}

func (u Network) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name: "network",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: ""},
		},
		Usage: "Logical network",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all network",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add a network",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file"},
					&cli.StringFlag{Name: "provider"},
					&cli.StringFlag{Name: "address"},
					&cli.StringFlag{Name: "namespace"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a network",
				Aliases: []string{"rm"},
				Action:  u.Remove,
			},
			{
				Name:    "save",
				Usage:   "Save a network",
				Aliases: []string{"sa"},
				Action:  u.Save,
			},
			Access{}.Commands(),
			ClientQoS{}.Commands(),
			VPNClient{}.Commands(),
			OpenVPN{}.Commands(),
			Output{}.Commands(),
			PrefixRoute{}.Commands(),
			Link{}.Commands(),
			FindHop{}.Commands(),
			SNAT{}.Commands(),
			DNAT{}.Commands(),
		},
	})
}

type SNAT struct {
	Cmd
}

func (s SNAT) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/snat"
}

func (s SNAT) Enable(c *cli.Context) error {
	network := c.String("name")
	url := s.Url(c.String("url"), network)

	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (s SNAT) Disable(c *cli.Context) error {
	network := c.String("name")
	url := s.Url(c.String("url"), network)

	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (s SNAT) Commands() *cli.Command {
	return &cli.Command{
		Name:  "snat",
		Usage: "Configure SNAT",
		Subcommands: []*cli.Command{
			{
				Name:    "enable",
				Usage:   "Enable snat",
				Aliases: []string{"en"},
				Action:  s.Enable,
			},
			{
				Name:    "disable",
				Usage:   "Disable snat",
				Aliases: []string{"dis"},
				Action:  s.Disable,
			},
		},
	}
}

type DNAT struct {
	Cmd
}

func (s DNAT) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/dnat"
}

func (s DNAT) Add(c *cli.Context) error {
	url := s.Url(c.String("url"), c.String("name"))
	value := &schema.DNAT{
		Protocol: c.String("protocol"),
		Dport:    c.Int("dport"),
		ToDport:  c.Int("todport"),
		Dest:     c.String("dest"),
		ToDest:   c.String("todest"),
	}
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, value, nil); err != nil {
		return err
	}

	return nil
}

func (s DNAT) Delete(c *cli.Context) error {
	url := s.Url(c.String("url"), c.String("name"))
	value := &schema.DNAT{
		Protocol: c.String("protocol"),
		Dport:    c.Int("dport"),
		ToDport:  c.Int("todport"),
		Dest:     c.String("dest"),
		ToDest:   c.String("todest"),
	}
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, value, nil); err != nil {
		return err
	}

	return nil
}

func (s DNAT) List(c *cli.Context) error {
	url := s.Url(c.String("url"), c.String("name"))
	clt := s.NewHttp(c.String("token"))

	var items []schema.DNAT
	if err := clt.GetJSON(url, &items); err == nil {
		return s.Out(items, c.String("format"), "")
	} else {
		return err
	}

}

func (s DNAT) Commands() *cli.Command {
	return &cli.Command{
		Name:  "dnat",
		Usage: "Configure DNAT",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Liat all dnat",
				Aliases: []string{"ls"},
				Action:  s.List,
			},
			{
				Name:  "add",
				Usage: "Add a dnat",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "protocol", Value: "tcp"},
					&cli.IntFlag{Name: "dport", Required: true},
					&cli.StringFlag{Name: "dest"},
					&cli.StringFlag{Name: "todest", Required: true},
					&cli.IntFlag{Name: "todport", Required: true},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a snat",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "protocol", Value: "tcp"},
					&cli.IntFlag{Name: "dport", Required: true},
					&cli.StringFlag{Name: "dest"},
				},
				Action: s.Delete,
			},
		},
	}
}

type PrefixRoute struct {
	Cmd
}

func (r PrefixRoute) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/route"
}

func (r PrefixRoute) Add(c *cli.Context) error {
	network := c.String("name")
	if len(network) == 0 {
		return libol.NewErr("invalid network")
	}
	pr := &schema.PrefixRoute{
		Prefix:  c.String("prefix"),
		NextHop: c.String("nexthop"),
		FindHop: c.String("findhop"),
		Metric:  c.Int("metric"),
	}
	url := r.Url(c.String("url"), network)
	clt := r.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, pr, nil); err != nil {
		return err
	}
	return nil
}

func (r PrefixRoute) Remove(c *cli.Context) error {
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

func (r PrefixRoute) Save(c *cli.Context) error {
	network := c.String("name")
	url := r.Url(c.String("url"), network)

	clt := r.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (r PrefixRoute) Tmpl() string {
	return `# total {{ len . }}
{{ps -25 "prefix"}} {{ps -25 "nexthop"}} {{ps -8 "metric"}}
{{- range . }}
{{ps -25 .Prefix}} {{ if .FindHop }}{{ps -25 .FindHop}}{{ else }}{{ps -25 .NextHop}}{{ end }} {{pi -8 .Metric }}
{{- end }}
`
}

func (r PrefixRoute) List(c *cli.Context) error {
	url := r.Url(c.String("url"), c.String("name"))
	clt := r.NewHttp(c.String("token"))
	var items []schema.PrefixRoute
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return r.Out(items, c.String("format"), r.Tmpl())
}

func (r PrefixRoute) Commands() *cli.Command {
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

type ClientQoS struct {
	Cmd
}

func (q ClientQoS) Commands() *cli.Command {
	return &cli.Command{
		Name:  "qos",
		Usage: "QoS for client",
		Subcommands: []*cli.Command{
			QosRule{}.Commands(),
		},
	}
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
		Name:    c.String("client"),
		InSpeed: c.Float64("inspeed"),
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
		Name: c.String("client"),
	}

	clt := qr.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, rule, nil); err != nil {
		return err
	}

	return nil
}

func (qr QosRule) Tmpl() string {
	return `# total {{ len . }}
{{ps -28 "Name"}} {{ps -10 "Device"}} {{ps -15 "Ip"}} {{ps -8 "InSpeed"}}
{{- range . }}
{{ps -28 .Name}} {{ps -10 .Device}} {{ps -15 .Ip}} {{pf -8 2 .InSpeed}}
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
					&cli.StringFlag{Name: "client", Aliases: []string{"c"}},
					&cli.Float64Flag{Name: "inspeed", Aliases: []string{"is"}},
				},
				Action: qr.Add,
			},
			{
				Name:    "remove",
				Usage:   "remove a qos rule",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "client", Aliases: []string{"c"}},
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
