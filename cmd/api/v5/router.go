package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

// openlan router tunnel add --remote 1.1.1.2 --address 192.168.1.1 --protocol gre
// openlan router tunnel add --remote 1.1.1.2 --address 192.168.1.1 --protocol ipip
// openlan router tunnel remove --remote 1.1.1.2 --address 192.168.1.1

type Router struct {
	Cmd
}

func (b Router) Url(prefix string) string {
	return prefix + "/api/network/router"
}

func (b Router) List(c *cli.Context) error {
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	var data schema.Network
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return b.Out(data, "yaml", "")
}

func (b Router) Save(c *cli.Context) error {
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (b Router) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "router",
		Usage:  "Special Router network",
		Action: b.List,
		Subcommands: []*cli.Command{
			{
				Name:    "ls",
				Usage:   "Display router network",
				Aliases: []string{"ls"},
				Action:  b.List,
			},
			{
				Name:    "save",
				Usage:   "Save router network",
				Aliases: []string{"sa"},
				Action:  b.Save,
			},
			RouterTunnel{}.Commands(),
			RouterPrivate{}.Commands(),
			RouterInterface{}.Commands(),
			KernelRoute{}.Commands(),
			KernelNeighbor{}.Commands(),
			Redirect{}.Commands(),
		},
	})
}

type RouterTunnel struct {
	Cmd
}

func (s RouterTunnel) Url(prefix string) string {
	return prefix + "/api/network/router/tunnel"
}

func (s RouterTunnel) Add(c *cli.Context) error {
	data := &schema.RouterTunnel{
		Remote:   c.String("remote"),
		Address:  c.String("address"),
		Protocol: c.String("protocol"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterTunnel) Remove(c *cli.Context) error {
	data := &schema.RouterTunnel{
		Remote:   c.String("remote"),
		Protocol: c.String("protocol"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterTunnel) Commands() *cli.Command {
	return &cli.Command{
		Name:  "tunnel",
		Usage: "Router tunnels",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add router tunnel",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "address", Required: true},
					&cli.StringFlag{Name: "remote", Required: true},
					&cli.StringFlag{Name: "protocol", Value: "gre"},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove router tunnel",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "remote", Required: true},
					&cli.StringFlag{Name: "protocol", Value: "gre"},
				},
				Action: s.Remove,
			},
		},
	}
}

type RouterPrivate struct {
	Cmd
}

func (s RouterPrivate) Url(prefix string) string {
	return prefix + "/api/network/router/private"
}

func (s RouterPrivate) Add(c *cli.Context) error {
	data := &schema.RouterPrivate{
		Subnet: c.String("subnet"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterPrivate) Remove(c *cli.Context) error {
	data := &schema.RouterPrivate{
		Subnet: c.String("subnet"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterPrivate) Commands() *cli.Command {
	return &cli.Command{
		Name:  "private",
		Usage: "Router private subnet",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add private",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "subnet", Required: true},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove private",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "subnet", Required: true},
				},
				Action: s.Remove,
			},
		},
	}
}

type RouterInterface struct {
	Cmd
}

func (s RouterInterface) Tmpl() string {
	return `# total {{ len . }}
{{ps -15 "name"}} {{ps -18 "address"}} {{ps -13 "mtu"}} {{ps -18 "mac"}} {{ps -22 "Statistics"}} {{ps -8 "Speed"}}
{{- range . }}
{{ps -15 .Name}} {{ps -18 .Address}} {{pi -13 .Mtu}} {{ps -18 .Mac}} {{pi 10 .Recv}}/{{pi -10 .Send}}/{{pi -2 .Drop}} {{pb .RxSpeed}}/{{pb .TxSpeed}}
{{- end }}
`
}

func (s RouterInterface) Url(prefix string) string {
	return prefix + "/api/network/router/interface"
}

func (s RouterInterface) List(c *cli.Context) error {
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	var data []schema.Device
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return s.Out(data, c.String("format"), s.Tmpl())
}

func (s RouterInterface) Add(c *cli.Context) error {
	data := &schema.RouterInterface{
		Device:  c.String("device"),
		VLAN:    c.Int("vlan"),
		Address: c.String("address"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterInterface) Remove(c *cli.Context) error {
	data := &schema.RouterInterface{
		Device: c.String("device"),
		VLAN:   c.Int("vlan"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s RouterInterface) Commands() *cli.Command {
	return &cli.Command{
		Name:   "interface",
		Usage:  "Router interface",
		Action: s.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List interface",
				Action: s.List,
			},
			{
				Name:  "add",
				Usage: "Add interface",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
					&cli.IntFlag{Name: "vlan"},
					&cli.StringFlag{Name: "address", Required: true},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove interface",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
					&cli.IntFlag{Name: "vlan"},
				},
				Action: s.Remove,
			},
		},
	}
}

type KernelRoute struct {
	Cmd
}

func (s KernelRoute) Tmpl() string {
	return `# total {{ len . }}
{{ps -18 "destination"}} {{ps -15 "nexthop"}} {{ps -15 "device"}} {{ps -6 "protocl"}} {{ps -15 "source"}} {{ ps -6 "metric"}}
{{- range . }}
{{ps -18 .Prefix}} {{ps -15 .NextHop}} {{ps -15 .Link}} {{ps -6 .Protocol}} {{ps -15 .Source}} {{pi -6 .Metric}}
{{- end }}
`
}

func (s KernelRoute) Url(prefix string) string {
	return prefix + "/api/kernel/route"
}

func (s KernelRoute) List(c *cli.Context) error {
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	var data []schema.KernelRoute
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return s.Out(data, c.String("format"), s.Tmpl())
}

func (s KernelRoute) Commands() *cli.Command {
	return &cli.Command{
		Name:   "route",
		Usage:  "Kernel routes",
		Action: s.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List kernel routes",
				Action: s.List,
			},
		},
	}
}

type KernelNeighbor struct {
	Cmd
}

func (s KernelNeighbor) Tmpl() string {
	return `# total {{ len . }}
{{ps -15 "address"}} {{ps -18 "mac"}} {{ps -15 "device"}}
{{- range . }}
{{ps -15 .Address}} {{ps -18 .HwAddr}} {{ps -15 .Link}}
{{- end }}
`
}

func (s KernelNeighbor) Url(prefix string) string {
	return prefix + "/api/kernel/neighbor"
}

func (s KernelNeighbor) List(c *cli.Context) error {
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	var data []schema.KernelNeighbor
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return s.Out(data, c.String("format"), s.Tmpl())
}

func (s KernelNeighbor) Commands() *cli.Command {
	return &cli.Command{
		Name:   "neighbor",
		Usage:  "Kernel neighbors",
		Action: s.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List kernel neighbors",
				Action: s.List,
			},
		},
	}
}

type Redirect struct {
	Cmd
}

func (s Redirect) Url(prefix string) string {
	return prefix + "/api/network/router/redirect"
}

func (s Redirect) Add(c *cli.Context) error {
	url := s.Url(c.String("url"))
	value := schema.RedirectRoute{
		Source:  c.String("source"),
		NextHop: c.String("nexthop"),
		Table:   c.Int("table"),
	}
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, &value, nil); err != nil {
		return err
	}
	return nil
}

func (s Redirect) Remove(c *cli.Context) error {
	url := s.Url(c.String("url"))
	value := schema.RedirectRoute{
		Source:  c.String("source"),
		NextHop: c.String("nexthop"),
		Table:   c.Int("table"),
	}
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, &value, nil); err != nil {
		return err
	}
	return nil
}

func (s Redirect) Commands() *cli.Command {
	return &cli.Command{
		Name:  "redirect",
		Usage: "Redirect route",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add redirect",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "source", Required: true},
					&cli.StringFlag{Name: "nexthop", Required: true},
					&cli.IntFlag{Name: "table", Required: true},
				},
				Action: s.Add,
			},
			{
				Name:  "remove",
				Usage: "Remove redirect",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "source", Required: true},
					&cli.StringFlag{Name: "nexthop", Required: true},
					&cli.IntFlag{Name: "table", Required: true},
				},
				Action: s.Remove,
			},
		},
	}
}
