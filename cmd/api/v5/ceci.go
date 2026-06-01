package v5

import (
	"errors"
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Ceci struct {
	Cmd
}

func (u Ceci) Url(prefix string) string {
	return prefix + "/api/network/ceci"
}

func (u Ceci) List(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	var data schema.Network
	if err := clt.GetJSON(url, &data); err != nil {
		return err
	}
	return u.Out(data, "yaml", "")
}

func (u Ceci) Save(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u Ceci) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "ceci",
		Usage:  "Special Ceci proxy network",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List a Ceci proxy",
				Action: u.List,
			},
			{
				Name:    "save",
				Usage:   "Save Ceci network",
				Aliases: []string{"sa"},
				Action:  u.Save,
			},
			CeciProxy{}.Commands(app),
			CeciService{}.Commands(app),
		},
	})
}

type CeciProxy struct {
	Cmd
}

type CeciService struct {
	Cmd
}
type CeciServiceHTTP struct {
	Cmd
}

func (u CeciService) Url(prefix string) string {
	return prefix + "/api/network/ceci/service"
}

func (u CeciService) Add(c *cli.Context) error {
	data := &schema.CeciProxy{
		Mode:    "service",
		Listen:  c.String("listen"),
		Network: c.String("network"),
		Service: &schema.CeciService{
			Protocol: c.String("protocol"),
			Balance:  c.String("balance"),
		},
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	return clt.PostJSON(url, data, nil)
}

func (u CeciService) addWithProtocol(c *cli.Context, protocol string) error {
	if protocol != "" && c.String("protocol") == "" {
		_ = c.Set("protocol", protocol)
	}
	return u.Add(c)
}

func (u CeciService) List(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	items := make([]schema.CeciProxy, 0, 16)
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, "yaml", "")
}

func (u CeciService) Remove(c *cli.Context) error {
	data := &schema.CeciProxy{Listen: c.String("listen")}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	return clt.DeleteJSON(url, data, nil)
}

func (u CeciService) Restart(c *cli.Context) error {
	data := &schema.CeciProxy{Listen: c.String("listen")}
	url := u.Url(c.String("url")) + "/restart"
	clt := u.NewHttp(c.String("token"))
	return clt.PutJSON(url, data, nil)
}

func (u CeciService) BackendAdd(c *cli.Context) error {
	listen := strings.TrimSpace(c.String("listen"))
	if listen == "" {
		return errors.New("listen is required")
	}
	data := &schema.CeciServiceBackendAdd{
		Listen:   listen,
		Hostname: strings.TrimSpace(c.String("hostname")),
		Backends: nil,
	}
	rawBackend := strings.TrimSpace(c.String("backend"))
	if rawBackend != "" {
		for _, item := range strings.FieldsFunc(rawBackend, func(r rune) bool { return r == '|' || r == ';' || r == ',' }) {
			item = strings.TrimSpace(item)
			if item != "" {
				data.Backends = append(data.Backends, item)
			}
		}
	}
	if data.Hostname != "" && len(data.Backends) == 0 {
		return errors.New("backend is required when hostname is set")
	}
	if data.Hostname == "" && len(data.Backends) == 0 {
		return errors.New("backend is required when hostname is empty")
	}
	url := u.Url(c.String("url")) + "/backend"
	clt := u.NewHttp(c.String("token"))
	return clt.PostJSON(url, data, nil)
}

func (u CeciService) Commands(app *api.App) *cli.Command {
	return &cli.Command{
		Name:   "service",
		Usage:  "Special Ceci service",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List Ceci services",
				Action: u.List,
			},
			{
				Name:  "add",
				Usage: "Add a Ceci Service",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
					&cli.StringFlag{Name: "network"},
					&cli.StringFlag{Name: "protocol"},
					&cli.StringFlag{Name: "balance"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a Ceci Service",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
				},
				Action: u.Remove,
			},
			{
				Name:  "restart",
				Usage: "Restart a Ceci Service",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
				},
				Action: u.Restart,
			},
			{
				Name:  "backend",
				Usage: "Manage service backends",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "Append backend, use --hostname with --backend for route backend",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "listen", Required: true},
							&cli.StringFlag{Name: "hostname"},
							&cli.StringFlag{Name: "backend"},
						},
						Action: u.BackendAdd,
					},
				},
			},
			CeciServiceHTTP{}.Commands(app),
		},
	}
}

func (u CeciServiceHTTP) Add(c *cli.Context) error {
	return CeciService{Cmd: u.Cmd}.addWithProtocol(c, "http")
}

func (u CeciServiceHTTP) BackendAdd(c *cli.Context) error {
	return CeciService{Cmd: u.Cmd}.BackendAdd(c)
}

func (u CeciServiceHTTP) Commands(app *api.App) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: "Ceci HTTP service",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a Ceci HTTP service",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
					&cli.StringFlag{Name: "network"},
					&cli.StringFlag{Name: "balance"},
				},
				Action: u.Add,
			},
			{
				Name:  "backend",
				Usage: "Manage HTTP service backend routes",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "Append one hostname route, use --hostname <host> --backend <server1|server2>",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "listen", Required: true},
							&cli.StringFlag{Name: "hostname"},
							&cli.StringFlag{Name: "backend"},
						},
						Action: u.BackendAdd,
					},
				},
			},
		},
	}
}

func (u CeciProxy) Url(prefix string) string {
	return prefix + "/api/network/ceci/proxy"
}

func (u CeciProxy) Add(c *cli.Context) error {
	target := strings.Split(c.String("target"), ",")
	data := &schema.CeciProxy{
		Mode:    c.String("mode"),
		Listen:  c.String("listen"),
		Network: c.String("network"),
		Target:  target,
	}
	if cert := c.String("cert"); cert != "" || c.String("key") != "" || c.String("root-ca") != "" || c.Bool("insecure") {
		data.Cert = &schema.Cert{
			CrtFile:  cert,
			KeyFile:  c.String("key"),
			CaFile:   c.String("root-ca"),
			Insecure: c.Bool("insecure"),
		}
	}
	if data.Mode != "tcp" && data.Mode != "http" && data.Mode != "name" {
		return errors.New("invalid mode, must be 'tcp', 'http' or 'name'")
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u CeciProxy) Remove(c *cli.Context) error {
	data := &schema.CeciProxy{
		Listen: c.String("listen"),
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u CeciProxy) Restart(c *cli.Context) error {
	data := &schema.CeciProxy{
		Listen: c.String("listen"),
	}
	url := u.Url(c.String("url")) + "/restart"
	clt := u.NewHttp(c.String("token"))
	if err := clt.PutJSON(url, data, nil); err != nil {
		return err
	}
	return nil
}

func (u CeciProxy) List(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	items := make([]schema.CeciProxy, 0, 16)
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	u.Out(items, "yaml", "")
	return nil
}

func (u CeciProxy) Commands(app *api.App) *cli.Command {
	return &cli.Command{
		Name:   "proxy",
		Usage:  "Special Ceci proxy",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:   "ls",
				Usage:  "List Ceci Proxy",
				Action: u.List,
			},
			{
				Name:  "add",
				Usage: "Add a Ceci Proxy",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
					&cli.StringFlag{Name: "mode", Required: true},
					&cli.StringFlag{Name: "network"},
					&cli.StringFlag{Name: "target"},
					&cli.StringFlag{Name: "cert"},
					&cli.StringFlag{Name: "key"},
					&cli.StringFlag{Name: "root-ca"},
					&cli.BoolFlag{Name: "insecure"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a Ceci Proxy",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
				},
				Action: u.Remove,
			},
			{
				Name:  "restart",
				Usage: "Restart a Ceci Proxy",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Required: true},
				},
				Action: u.Restart,
			},
		},
	}
}
