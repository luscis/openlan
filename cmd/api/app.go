package api

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
)

const (
	ConfSockFile   = "unix:/var/openlan/confd/confd.sock"
	ConfDatabase   = "OpenLAN_Switch"
	AdminTokenFile = "/etc/openlan/switch/token"
)

var (
	Version  = "v5"
	Url      = "https://localhost:10000"
	Token    = ""
	Server   = ConfSockFile
	Database = ConfDatabase
	Verbose  = false
)

type App struct {
	cli    *cli.App
	Before func(c *cli.Context) error
	After  func(c *cli.Context) error
}

func (a *App) Flags() []cli.Flag {
	var flags []cli.Flag

	switch Version {
	case "v6":
		flags = append(flags,
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "output format: json|yaml",
				Value:   "yaml",
			})
		flags = append(flags,
			&cli.StringFlag{
				Name:    "conf",
				Aliases: []string{"c"},
				Usage:   "confd server connection",
				Value:   Server,
			})
		flags = append(flags,
			&cli.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Usage:   "confd database",
				Value:   Database,
			})
	default:
		flags = append(flags,
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "output format: json|yaml",
				Value:   "table",
			})
		flags = append(flags,
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "admin token",
				Value:   Token,
			})
		flags = append(flags,
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"l"},
				Usage:   "server url",
				Value:   Url,
			})
	}
	flags = append(flags,
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "enable verbose",
			Value:   false,
		})
	return flags
}

func (a *App) New() *cli.App {
	app := &cli.App{
		Usage:    "OpenLAN switch utility",
		Flags:    a.Flags(),
		Commands: []*cli.Command{},
		Before: func(c *cli.Context) error {
			if c.Bool("verbose") {
				Verbose = true
				libol.SetLogger("", libol.DEBUG)
			} else {
				Verbose = false
				libol.SetLogger("", libol.INFO)
			}
			if a.Before == nil {
				return nil
			}
			return a.Before(c)
		},
		After: func(c *cli.Context) error {
			if a.After == nil {
				return nil
			}
			return a.After(c)
		},
	}
	a.cli = app
	return a.cli
}

func (a *App) Command(cmd *cli.Command) {
	a.cli.Commands = append(a.cli.Commands, cmd)
}

func (a *App) Run(args []string) error {
	return a.cli.Run(args)
}
