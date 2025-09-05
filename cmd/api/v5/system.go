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
{{ps -18 "destination"}} {{ps -15 "nexthop"}} {{ps -16 "link"}} {{ps -15 "source"}} {{"metric"}}
{{- range . }}
{{ps -18 .Prefix}} {{ps -15 .NextHop}} {{ps -16 .Link}} {{ps -15 .Source}} {{.Metric}}
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
		Name:  "prefix",
		Usage: "System prefix",
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
		Name:    "log",
		Aliases: []string{"v"},
		Usage:   "show log information",
		Action:  v.List,
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
{{ps -13 "name"}} {{ps -13 "mtu"}} {{ps -16 "mac"}} {{ps -6 "provider"}}
{{- range . }}
{{ps -13 .Name}} {{pi -13 .Mtu}} {{ps -16 .Mac}} {{ps -6 .Provider}}
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
		Name:  "device",
		Usage: "linux network device",
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
