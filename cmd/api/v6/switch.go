package v6

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/database"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
)

type Switch struct {
}

func (u Switch) List(c *cli.Context) error {
	var listSw []database.Switch
	if err := database.Client.List(&listSw); err == nil {
		return api.Out(listSw, c.String("format"), "")
	}
	return nil
}

func (u Switch) Add(c *cli.Context) error {
	protocol := c.String("protocol")
	listen := c.Int("listen")
	newSw := database.Switch{
		Protocol: protocol,
		Listen:   listen,
	}
	sw, _ := database.Client.Switch()
	if sw == nil {
		ops, err := database.Client.Create(&newSw)
		if err != nil {
			return err
		}
		libol.Debug("Switch.Add %s", ops)
		if ret, err := database.Client.Transact(ops...); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	} else {
		ops, err := database.Client.Where(sw).Update(&newSw)
		if err != nil {
			return err
		}
		libol.Debug("Switch.Add %s", ops)
		if ret, err := database.Client.Transact(ops...); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	}
	return nil
}

func (u Switch) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "switch",
		Aliases: []string{"sw"},
		Usage:   "Global switch",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List global switch",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add or update switch",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "protocol",
						Value: "tcp",
						Usage: "used protocol: tcp|udp|http|tls"},
					&cli.IntFlag{
						Name:  "listen",
						Value: 10002,
						Usage: "listen on port: 1024-65535",
					},
				},
				Action: u.Add,
			},
		},
	})
}
