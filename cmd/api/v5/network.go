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
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
				Action: u.List,
			},
			{
				Name:  "add",
				Usage: "Add a network",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file"},
					&cli.StringFlag{Name: "name"},
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
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
				Action: u.Remove,
			},
			{
				Name:    "save",
				Usage:   "Save a network",
				Aliases: []string{"sa"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
				Action: u.Save,
			},
			Access{}.Commands(),
			Qos{}.Commands(),
			VPNClient{}.Commands(),
			OpenVPN{}.Commands(),
			Output{}.Commands(),
			Route{}.Commands(),
			Link{}.Commands(),
			FindHop{}.Commands(),
			SNAT{}.Commands(),
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
		Usage: "Control SNAT",
		Subcommands: []*cli.Command{
			{
				Name:   "enable",
				Usage:  "enable snat",
				Action: s.Enable,
			},
			{
				Name:   "disable",
				Usage:  "disable snat",
				Action: s.Disable,
			},
		},
	}
}
