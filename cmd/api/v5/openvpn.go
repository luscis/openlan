package v5

import (
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type VPNClient struct {
	Cmd
}

func (u VPNClient) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/vpn/client"
	} else {
		return prefix + "/api/vpn/client/" + name
	}
}

func (u VPNClient) Tmpl() string {
	return `# total {{ len . }}
{{ps -8 "alive"}} {{ps -16 "address"}} {{ ps -13 "device" }} {{ps -15 "name"}} {{ps -22 "remote"}} {{ ps -6 "state"}}
{{- range . }}
{{pt .AliveTime | ps -8}} {{ps -16 .Address}} {{ ps -13 .Device }} {{ps -15 .Name}} {{ps -22 .Remote}} {{ ps -6 .State}}
{{- end }}
`
}

func (u VPNClient) List(c *cli.Context) error {
	url := u.Url(c.String("url"), c.String("name"))
	clt := u.NewHttp(c.String("token"))
	var items []schema.VPNClient
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	name := c.String("name")
	if len(name) > 0 {
		tmp := items[:0]
		for _, obj := range items {
			if obj.Network == name {
				tmp = append(tmp, obj)
			}
		}
		items = tmp
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u VPNClient) Commands() *cli.Command {
	return &cli.Command{
		Name:  "client",
		Usage: "Clients by OpenVPN",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all clients",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	}
}

type OpenVPN struct {
	Cmd
}

func (o OpenVPN) Url(prefix, name string) string {
	return prefix + "/api/network/" + name + "/openvpn/restart"
}

func (o OpenVPN) Restart(c *cli.Context) error {
	network := c.String("name")
	url := o.Url(c.String("url"), network)

	clt := o.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (o OpenVPN) Commands() *cli.Command {
	return &cli.Command{
		Name:  "openvpn",
		Usage: "Control OpenVPN",
		Subcommands: []*cli.Command{
			{
				Name:   "restart",
				Usage:  "restart openvpn for the network",
				Action: o.Restart,
			},
		},
	}
}
