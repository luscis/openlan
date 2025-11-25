package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

// openlan bgp enable --router-id 1.1.1.1 --local-as 31
// openlan bgp disable
//
// openlan bgp neighbor add --address 1.1.1.2 --remote-as 32
// openlan bgp neighbor add --address 1.1.1.3 --remote-as 33
//
// openlan bgp advertis add --neighbor 1.1.1.2 --prefix 192.168.1.0/24
// openlan bgp receives add --neighbor 1.1.1.2 --prefix 192.168.2.0/24

type BGP struct {
	Cmd
}

func (b BGP) Url(prefix string) string {
	return prefix + "/api/network/bgp/global"
}

func (b BGP) List(c *cli.Context) error {
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	var data schema.Bgp
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return b.Out(data, "yaml", "")
}

func (b BGP) Save(c *cli.Context) error {
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (b BGP) Enable(c *cli.Context) error {
	data := &schema.Bgp{
		LocalAs:  c.Int("local-as"),
		RouterId: c.String("router-id"),
	}
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (b BGP) Disable(c *cli.Context) error {
	url := b.Url(c.String("url"))
	clt := b.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (b BGP) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "bgp",
		Usage: "External BGP",
		Subcommands: []*cli.Command{
			{
				Name:    "ls",
				Usage:   "Display bgp",
				Aliases: []string{"ls"},
				Action:  b.List,
			},
			{
				Name:  "enable",
				Usage: "Enable bgp",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "router-id", Required: true},
					&cli.IntFlag{Name: "local-as", Required: true},
				},
				Action: b.Enable,
			},
			{
				Name:   "disable",
				Usage:  "Disable bgp",
				Action: b.Disable,
			},
			{
				Name:    "save",
				Usage:   "Save bgp network",
				Aliases: []string{"sa"},
				Action:  b.Save,
			},
			Neighbor{}.Commands(),
			Advertis{}.Commands(),
			Receives{}.Commands(),
		},
	})
}

type Neighbor struct {
	Cmd
}

func (s Neighbor) Url(prefix string) string {
	return prefix + "/api/network/bgp/neighbor"
}

func (s Neighbor) Add(c *cli.Context) error {
	data := &schema.BgpNeighbor{
		RemoteAs: c.Int("remote-as"),
		Address:  c.String("address"),
		Password: c.String("password"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s Neighbor) Remove(c *cli.Context) error {
	data := &schema.BgpNeighbor{
		Address: c.String("address"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s Neighbor) Commands() *cli.Command {
	return &cli.Command{
		Name:  "neighbor",
		Usage: "BGP neighbor",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add BGP neighbor",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "address", Required: true},
					&cli.IntFlag{Name: "remote-as", Required: true},
					&cli.StringFlag{Name: "password"},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove BGP neighbor",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "address", Required: true},
				},
				Action: s.Remove,
			},
		},
	}
}

type Advertis struct {
	Cmd
}

func (s Advertis) Url(prefix string) string {
	return prefix + "/api/network/bgp/advertis"
}

func (s Advertis) Add(c *cli.Context) error {
	data := &schema.BgpPrefix{
		Prefix:   c.String("prefix"),
		Neighbor: c.String("neighbor"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s Advertis) Remove(c *cli.Context) error {
	data := &schema.BgpPrefix{
		Prefix:   c.String("prefix"),
		Neighbor: c.String("neighbor"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s Advertis) Commands() *cli.Command {
	return &cli.Command{
		Name:  "advertis",
		Usage: "Neighbor advertised routes",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add advertis prefix",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "neighbor", Required: true},
					&cli.StringFlag{Name: "prefix", Required: true},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove advertis prefix",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "neighbor", Required: true},
					&cli.StringFlag{Name: "prefix", Required: true},
				},
				Action: s.Remove,
			},
		},
	}
}

type Receives struct {
	Cmd
}

func (s Receives) Url(prefix string) string {
	return prefix + "/api/network/bgp/receives"
}

func (s Receives) Add(c *cli.Context) error {
	data := &schema.BgpPrefix{
		Prefix:   c.String("prefix"),
		Neighbor: c.String("neighbor"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s Receives) Remove(c *cli.Context) error {
	data := &schema.BgpPrefix{
		Prefix:   c.String("prefix"),
		Neighbor: c.String("neighbor"),
	}
	url := s.Url(c.String("url"))
	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (s Receives) Commands() *cli.Command {
	return &cli.Command{
		Name:  "receives",
		Usage: "Neighbor received prefix",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add received prefix",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "neighbor", Required: true},
					&cli.StringFlag{Name: "prefix", Required: true},
				},
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove received prefix",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "neighbor", Required: true},
					&cli.StringFlag{Name: "prefix", Required: true},
				},
				Action: s.Remove,
			},
		},
	}
}
