package v5

import (
	"strings"

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
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u VPNClient) Add(c *cli.Context) error {
	url := u.Url(c.String("url"), c.String("name"))

	value := &schema.VPNClient{
		Name:    c.String("user"),
		Address: c.String("address"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, value, nil); err != nil {
		return err
	}

	return nil
}

func (u VPNClient) Remove(c *cli.Context) error {
	url := u.Url(c.String("url"), c.String("name"))

	value := &schema.VPNClient{
		Name: c.String("user"),
	}

	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, value, nil); err != nil {
		return err
	}

	return nil
}

func (u VPNClient) Commands() *cli.Command {
	return &cli.Command{
		Name:  "client",
		Usage: "OpenVPN's client",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all clients",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add a client",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Required: true},
					&cli.StringFlag{Name: "address", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a client",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Required: true},
				},
				Action: u.Remove,
			},
		},
	}
}

type OpenVPN struct {
	Cmd
}

func (o OpenVPN) Url(prefix, name, action string) string {
	if action != "" {
		return prefix + "/api/network/" + name + "/openvpn/" + action
	}
	return prefix + "/api/network/" + name + "/openvpn"
}

func (o OpenVPN) Add(c *cli.Context) error {
	network := c.String("name")
	url := o.Url(c.String("url"), network, "")

	data := schema.OpenVPN{
		Listen:   c.String("listen"),
		Protocol: c.String("protocol"),
		Subnet:   c.String("subnet"),
	}
	dns := strings.Split(c.String("dns"), ",")
	for _, v := range dns {
		data.Push = append(data.Push, "dhcp-option DNS "+v)
	}
	clt := o.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, &data, nil); err != nil {
		return err
	}
	return nil
}

func (o OpenVPN) Remove(c *cli.Context) error {
	network := c.String("name")
	url := o.Url(c.String("url"), network, "")
	clt := o.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (o OpenVPN) Restart(c *cli.Context) error {
	network := c.String("name")
	url := o.Url(c.String("url"), network, "restart")
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
				Name:  "add",
				Usage: "Add openvpn",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "protocol", Value: "tcp"},
					&cli.StringFlag{Name: "listen", Required: true},
					&cli.StringFlag{Name: "subnet"},
					&cli.StringFlag{Name: "dns"},
				},
				Action: o.Add,
			},
			{
				Name:   "remove",
				Usage:  "Remove openvpn",
				Action: o.Remove,
			},
			{
				Name:   "restart",
				Usage:  "Restart openvpn",
				Action: o.Restart,
			},
		},
	}
}
