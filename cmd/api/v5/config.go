package v5

import (
	"fmt"
	"path/filepath"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Cmd
}

func (u Config) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/config"
	}
	return prefix + "/api/config/" + name
}

func (u Config) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	cfg := &config.Switch{}
	if err := clt.GetJSON(url, cfg); err == nil {
		name := c.String("network")
		format := c.String("format")
		if format == "yaml" {
			cfg.FormatNetwork()
		}
		if len(name) > 0 {
			obj := cfg.GetNetwork(name)
			return u.Out(obj, format, "")
		} else {
			return u.Out(cfg, format, "")
		}
	} else {
		return err
	}
}

func (u Config) Check(c *cli.Context) error {
	out := u.Log()
	dir := c.String("dir")
	// Check proxy configurations.
	out.Info("%15s: %s", "check", "proxy")
	file := filepath.Join(dir, "proxy.json")
	if err := libol.FileExist(file); err == nil {
		obj := &config.Proxy{}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			out.Warn("%15s: %s", filepath.Base(file), err)
		} else {
			out.Info("%15s: %s", filepath.Base(file), "success")
		}
	}
	// Check OLAP configurations.
	out.Info("%15s: %s", "check", "point")
	file = filepath.Join(dir, "point.json")
	if err := libol.FileExist(file); err == nil {
		obj := &config.Point{}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			out.Warn("%15s: %s", filepath.Base(file), err)
		} else {
			out.Info("%15s: %s", filepath.Base(file), "success")
		}
	}
	// Check OLSW configurations.
	out.Info("%15s: %s", "check", "switch")
	file = filepath.Join(dir, "switch", "switch.json")
	if err := libol.FileExist(file); err == nil {
		obj := &config.Switch{}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			out.Warn("%15s: %s", filepath.Base(file), err)
		} else {
			out.Info("%15s: %s", filepath.Base(file), "success")
		}
	}
	// Check network configurations.
	out.Info("%15s: %s", "check", "network")
	pattern := filepath.Join(dir, "switch", "network", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			obj := &config.Network{}
			if err := libol.UnmarshalLoad(obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}

	out.Info("%15s: %s", "check", "qos")
	pattern = filepath.Join(dir, "switch", "qos", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			obj := &config.Qos{}
			if err := libol.UnmarshalLoad(obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}

	// Check ACL configurations.
	out.Info("%15s: %s", "check", "acl")
	pattern = filepath.Join(dir, "switch", "acl", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			obj := &config.ACL{}
			if err := libol.UnmarshalLoad(obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}
	// Check links configurations.
	out.Info("%15s: %s", "check", "link")
	pattern = filepath.Join(dir, "switch", "link", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			var obj []config.Point
			if err := libol.UnmarshalLoad(&obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}
	// Check routes configurations.
	out.Info("%15s: %s", "check", "route")
	pattern = filepath.Join(dir, "switch", "route", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			var obj []config.PrefixRoute
			if err := libol.UnmarshalLoad(&obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}

	// Check output config
	out.Info("%15s: %s", "check", "output")
	pattern = filepath.Join(dir, "switch", "output", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			var obj []config.Output
			if err := libol.UnmarshalLoad(&obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}

	return nil
}

func (u Config) Reload(c *cli.Context) error {
	url := u.Url(c.String("url"), "reload")
	clt := u.NewHttp(c.String("token"))
	data := &schema.Message{}
	if err := clt.PutJSON(url, nil, data); err == nil {
		fmt.Println(data.Message)
		return nil
	} else {
		return err
	}
}

func (u Config) Save(c *cli.Context) error {
	url := u.Url(c.String("url"), "save")
	clt := u.NewHttp(c.String("token"))
	data := &schema.Message{}
	if err := clt.PutJSON(url, nil, data); err == nil {
		fmt.Println(data.Message)
		return nil
	} else {
		return err
	}
}

func (u Config) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "config",
		Usage: "Switch configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all configuration",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Value: ""},
				},
				Action: u.List,
			},
			{
				Name:    "check",
				Usage:   "Check all configuration",
				Aliases: []string{"co"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dir", Value: "/etc/openlan"},
				},
				Action: u.Check,
			},
			{
				Name:    "reload",
				Usage:   "Reload configuration",
				Aliases: []string{"re"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dir", Value: "/etc/openlan"},
				},
				Action: u.Reload,
			},
			{
				Name:    "save",
				Usage:   "Save configuration",
				Aliases: []string{"sa"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dir", Value: "/etc/openlan"},
				},
				Action: u.Save,
			},
		},
	})
}
