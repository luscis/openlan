package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/urfave/cli/v2"
)

// openlan bgp enable --route-id 1.1.1.1 --as-path 30
// openlan bgp disable
// openlan bgp add neighbor --address 1.1.1.2 --remote-as 32

type BGP struct {
	Cmd
}

func (b BGP) Url(prefix string) string {
	return prefix + "/api/bgp"
}

func (b BGP) List(c *cli.Context) error {
	return nil
}

func (b BGP) Enable(c *cli.Context) error {
	return nil
}

func (b BGP) Disable(c *cli.Context) error {
	return nil
}

func (b BGP) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:  "bgp",
		Usage: "External BGP",
		Subcommands: []*cli.Command{
			{
				Name:    "ls",
				Usage:   "Display bgp",
				Aliases: []string{"ls"},
				Action:  b.List,
			},
			{
				Name:   "enable",
				Usage:  "Enable bgp",
				Action: b.Enable,
			},
			{
				Name:   "disable",
				Usage:  "Disable bgp",
				Action: b.Disable,
			},
			Neighbor{}.Commands(),
		},
	})
}

type Neighbor struct {
	Cmd
}

func (s Neighbor) Url(prefix string) string {
	return prefix + "/api/bgp/neighbor/"
}

func (s Neighbor) Add(c *cli.Context) error {
	url := s.Url(c.String("url"))

	clt := s.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (s Neighbor) Remove(c *cli.Context) error {
	url := s.Url(c.String("url"))

	clt := s.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil, nil); err != nil {
		return err
	}

	return nil
}

func (s Neighbor) Commands() *cli.Command {
	return &cli.Command{
		Name:  "neighbor",
		Usage: "BGP neighbor",
		Subcommands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Add BGP neighbor",
				Action: s.Add,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove BGP neighbor",
				Action:  s.Remove,
			},
		},
	}
}
