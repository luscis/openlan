package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Network struct {
	Cmd
}

func (u Network) Url(prefix, name string) string {
	return prefix + "/api/network"
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
