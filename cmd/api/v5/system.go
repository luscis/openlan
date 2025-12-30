package v5

import (
	"fmt"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Prefix struct {
	Cmd
}

func (r Prefix) Url(prefix string) string {
	return prefix + "/api/prefix"
}

func (r Prefix) Tmpl() string {
	return `# total {{ len . }}
{{ps -18 "destination"}} {{ps -15 "nexthop"}} {{ps -16 "link"}} {{ps -15 "source"}} {{ps -6 "metric" }} {{ "protocol" }}
{{- range . }}
{{ps -18 .Prefix}} {{ps -15 .NextHop}} {{ps -16 .Link}} {{ps -15 .Source}} {{pi -6 .Metric}} {{ .Protocol }}
{{- end }}
`
}

func (r Prefix) List(c *cli.Context) error {
	url := r.Url(c.String("url"))
	clt := r.NewHttp(c.String("token"))
	var items []schema.PrefixRoute
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return r.Out(items, c.String("format"), r.Tmpl())
}

func (r Prefix) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "prefix",
		Usage:  "System prefix",
		Action: r.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List system routes.",
				Action:  r.List,
			},
		},
	})
}

type Log struct {
	Cmd
}

func (v Log) Url(prefix, name string) string {
	return prefix + "/api/log"
}

func (v Log) Tmpl() string {
	return `File :  {{ .File }}
Level:  {{ .Level}}
`
}

func (v Log) List(c *cli.Context) error {
	url := v.Url(c.String("url"), "")
	clt := v.NewHttp(c.String("token"))
	var item schema.Log
	if err := clt.GetJSON(url, &item); err != nil {
		return err
	}
	return v.Out(item, c.String("format"), v.Tmpl())
}

func (v Log) Add(c *cli.Context) error {
	url := v.Url(c.String("url"), "")
	log := &schema.Log{
		Level: c.Int("level"),
	}
	clt := v.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, log, nil); err != nil {
		return err
	}
	return nil
}

func (v Log) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "log",
		Usage:  "show log information",
		Action: v.List,
		Subcommands: []*cli.Command{
			{
				Name:  "set",
				Usage: "set log level",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "level"},
				},
				Action: v.Add,
			},
		},
	})
}

type Device struct {
	Cmd
}

func (u Device) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/device"
	} else {
		return prefix + "/api/device/" + name
	}
}

func (u Device) Tmpl() string {
	return `# total {{ len . }}
{{ps -15 "name"}} {{ps -13 "mtu"}} {{ps -6 "provider"}} {{ps -16 ".Statistics"}}
{{- range . }}
{{ps -15 .Name}} {{pi -13 .Mtu}} {{ps -6 .Provider}} {{pi 8 .Recv}}/{{pi 8 .Send}}/{{pi 2 .Drop}}
{{- end }}
`
}

func (u Device) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Device
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Device) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "device",
		Usage:  "linux network device",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all devices",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}

type PProf struct {
	Cmd
}

func (u PProf) Url(prefix, name string) string {
	return prefix + "/api/pprof"
}

func (u PProf) Add(c *cli.Context) error {
	pp := schema.PProf{
		Listen: c.String("listen"),
	}
	if pp.Listen == "" {
		return libol.NewErr("listen value is empty")
	}
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, pp, nil); err != nil {
		return err
	}
	return nil
}

func (u PProf) Del(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u PProf) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var pp schema.PProf
	if err := clt.GetJSON(url, &pp); err != nil {
		return err
	}
	fmt.Println(pp.Listen)
	return nil
}

func (u PProf) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "pprof",
		Usage: "Debug pprof tool",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Show configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "enable",
				Usage: "Enable pprof tool",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "listen", Value: "127.0.0.1:6060"},
				},
				Action: u.Add,
			},
		},
	})
}

type RateLimit struct {
	Cmd
}

func (u RateLimit) Url(prefix, name string) string {
	return prefix + "/api/interface/" + name + "/rate"
}

func (u RateLimit) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "device"}} {{ps -8 "speed"}}
{{- range . }}
{{ps -16 .Device}} {{pi .Speed}}
{{- end }}
`
}

func (u RateLimit) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Rate
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u RateLimit) Add(c *cli.Context) error {
	name := c.String("device")
	rate := &schema.Rate{
		Device: name,
		Speed:  c.Int("speed"),
	}
	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, rate, nil); err != nil {
		return err
	}
	return nil
}

func (u RateLimit) Remove(c *cli.Context) error {
	name := c.String("device")

	url := u.Url(c.String("url"), name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}
	return nil
}

func (u RateLimit) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "rate",
		Usage: "Rate Limit",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a rate limit",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
					&cli.StringFlag{Name: "speed", Required: true},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove a rate limit",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "device", Required: true},
				},
				Action: u.Remove,
			},
		},
	})
}

type Server struct {
	Cmd
}

func (u Server) Url(prefix, name string) string {
	return prefix + "/api/server"
}

func (u Server) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	if data, err := clt.GetBody(url); err == nil {
		fmt.Println(string(data))
		return nil
	} else {
		return err
	}
}

func (u Server) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "server",
		Usage:  "Socket server status",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display server status",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}

type Ldap struct {
	Cmd
}

func (u Ldap) Url(prefix string) string {
	return prefix + "/api/ldap"
}

func (u Ldap) List(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	value := &schema.LDAP{}
	if err := clt.GetJSON(url, &value); err == nil {
		u.Out(value, "", "")
		return nil
	} else {
		return err
	}
}

func (u Ldap) Add(c *cli.Context) error {
	value := &schema.LDAP{
		Server:    c.String("server"),
		BindDN:    c.String("bindDN"),
		BindPass:  c.String("bindPass"),
		BaseDN:    c.String("baseDN"),
		Attribute: c.String("attribute"),
		Filter:    c.String("filter"),
		EnableTls: c.Bool("isTLS"),
	}
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, value, nil); err != nil {
		return err
	}
	return nil
}

func (u Ldap) Remove(c *cli.Context) error {
	url := u.Url(c.String("url"))
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, "", nil); err != nil {
		return err
	}
	return nil
}

func (u Ldap) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "ldap",
		Usage:  "Configure LDAP",
		Action: u.List,
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display ldap",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add ladp server",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "server", Required: true},
					&cli.StringFlag{Name: "bindDN", Required: true},
					&cli.StringFlag{Name: "bindPass", Required: true},
					&cli.StringFlag{Name: "baseDN", Required: true},
					&cli.StringFlag{Name: "attribute", Required: true},
					&cli.StringFlag{Name: "filter", Required: true},
					&cli.BoolFlag{Name: "isTLS", Value: false},
				},
				Action: u.Add,
			},
			{
				Name:   "remove",
				Usage:  "Remove ladp server",
				Action: u.Remove,
			},
		},
	})
}
