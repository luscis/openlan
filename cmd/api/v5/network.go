package v5

import (
	"github.com/luscis/openlan/cmd/api"
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
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Network
	if err := clt.GetJSON(url, &items); err == nil {
		return u.Out(items, c.String("format"), "")
	} else {
		return err
	}
}

func (u Network) Add(c *cli.Context) error {
	file := c.String("file")
	network := &schema.Network{}

	if err := libol.UnmarshalLoad(&network.Config, file); err != nil {
		return err
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
	if len(network) == 0 {
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
	openvpn := OpenVpn{}
	app.Command(&cli.Command{
		Name:    "network",
		Aliases: []string{"net"},
		Usage:   "Logical network",
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
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove the network",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
				Action: u.Remove,
			},
			{
				Name:    "save",
				Usage:   "Save the network",
				Aliases: []string{"sa"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Value: ""},
				},
				Action: u.Save,
			},
			openvpn.Commands(),
		},
	})
}

type OpenVpn struct {
	Cmd
}

func (o OpenVpn) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/openvpn/restart"
}

func (o OpenVpn) Restart(c *cli.Context) error {
	network := c.String("network")
	url := o.Url(c.String("url"), network)

	clt := o.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (o OpenVpn) Commands() *cli.Command {
	return &cli.Command{
		Name:    "openvpn",
		Usage:   "control openvpn",
		Aliases: []string{"ov"},
		Subcommands: []*cli.Command{
			{
				Name:    "restart",
				Usage:   "restart openvpn for the network",
				Aliases: []string{"ro"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Required: true},
				},
				Action: o.Restart,
			},
		},
	}
}
