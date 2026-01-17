package v5

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Index struct {
	Cmd
}

func (r Index) Url(prefix string) string {
	return prefix + "/api/index"
}

func (r Index) List(c *cli.Context) error {
	url := r.Url(c.String("url"))
	clt := r.NewHttp(c.String("token"))
	var value schema.Index
	if err := clt.GetJSON(url, &value); err != nil {
		return err
	}
	return r.Out(value, c.String("format"), r.Tmpl())
}

func (r Index) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "index",
		Usage:  "Display information",
		Action: r.List,
	})
}

type Reload struct {
	Cmd
}

const openPidFile = "/etc/openlan/switch/pid"
const maxWaitSec = 60

func showProcessInfo(pid int) {
	procDir := fmt.Sprintf("/proc/%d", pid)
	_, err := os.Stat(procDir)
	if err != nil {
		fmt.Printf("Process %d not existed\n", pid)
		return
	}

	cmdlinePath := filepath.Join(procDir, "cmdline")
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		fmt.Printf("Cann't read %d cmdline: %v\n", pid, err)
		return
	}

	cmd := strings.ReplaceAll(string(data), "\x00", " ")
	cmd = strings.TrimSpace(cmd)

	fmt.Printf("  PID   %d   CMD: %s\n", pid, cmd)
}

func readPid(file string) (int, error) {
	if v, err := os.ReadFile(file); err != nil {
		return 0, err
	} else {
		pidStr := strings.TrimSpace(string(v))
		return strconv.Atoi(pidStr)
	}
}

func (r Reload) Do(c *cli.Context) error {
	oldPid, err := readPid(openPidFile)
	if err != nil {
		return err
	}

	cfg := Config{}
	if err := cfg.Save(c); err != nil {
		return err
	}

	fmt.Printf("# reloading pid:%d ....\n", oldPid)
	showProcessInfo(oldPid)

	if proc, err := libol.Kill(oldPid); err != nil {
		return libol.NewErr("kill %d failed: %v", oldPid, err)
	} else {
		proc.Wait()
	}

	fmt.Printf("# max wait %ds...\n", maxWaitSec)

	last := time.Now().Unix()
	for range maxWaitSec {
		time.Sleep(1 * time.Second)
		newPid, err := readPid(openPidFile)
		if err != nil {
			return err
		}
		if newPid != oldPid {
			now := time.Now().Unix()
			fmt.Printf("# during %ds, new pid:%d ...\n", now-last, newPid)
			showProcessInfo(newPid)
			break
		}
		fmt.Printf("# ...\n")
	}
	return nil
}

func (r Reload) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:   "reload",
		Usage:  "Reload OpenLAN Switch",
		Action: r.Do,
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
		Usage:  "Show log information",
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

func (u Device) Url(prefix string) string {
	return prefix + "/api/device"

}

func (u Device) Tmpl() string {
	return `# total {{ len . }}
{{ps -15 "network"}} {{ps -15 "name"}} {{ps -13 "mtu"}} {{ps -18 "mac"}} {{ps -24 "Statistics"}} {{ps -8 "Speed"}}
{{- range . }}
{{ps -15 .Network}} {{ps -15 .Name}} {{pi -13 .Mtu}} {{ps -18 .Mac}} {{pi -10 .Recv}}/{{pi -10 .Send}}/{{pi -2 .Drop}} {{pb .RxSpeed}}/{{pb .TxSpeed}}
{{- end }}
`
}

func (u Device) List(c *cli.Context) error {
	url := u.Url(c.String("url"))
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
		Name:  "ratelimit",
		Usage: "Rate limit for device",
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
