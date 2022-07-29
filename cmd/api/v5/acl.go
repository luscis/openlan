package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/urfave/cli/v2"
)

type ACL struct {
	Cmd
}

func (u ACL) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/acl"
	} else {
		return prefix + "/api/acl/" + name
	}
}

func (u ACL) Add(c *cli.Context) error {
	return nil
}

func (u ACL) Remove(c *cli.Context) error {
	return nil
}

func (u ACL) List(c *cli.Context) error {
	return nil
}

func (u ACL) Apply(c *cli.Context) error {
	return nil
}

func (u ACL) Commands(app *api.App) {
	rule := ACLRule{}
	app.Command(&cli.Command{
		Name:  "acl",
		Usage: "Access control list",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}},
		},
		Subcommands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Add a new acl",
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing acl",
				Aliases: []string{"ls"},
				Action:  u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all acl",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			rule.Commands(),
			{
				Name:  "apply",
				Usage: "Apply a new acl",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Aliases: []string{"net"}},
				},
				Action: u.Apply,
			},
		},
	})
}

type ACLRule struct {
	Cmd
}

func (u ACLRule) Url(prefix, acl, name string) string {
	if name == "" {
		return prefix + "/api/acl/" + acl
	} else {
		return prefix + "/api/acl/" + acl + "/" + name
	}
}

func (u ACLRule) Add(c *cli.Context) error {
	return nil
}

func (u ACLRule) Remove(c *cli.Context) error {
	return nil
}

func (u ACLRule) List(c *cli.Context) error {
	return nil
}

func (u ACLRule) Commands() *cli.Command {
	return &cli.Command{
		Name:  "rule",
		Usage: "Access control list rule",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new acl rule",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "src", Aliases: []string{"s"}},
					&cli.StringFlag{Name: "dst", Aliases: []string{"d"}},
					&cli.StringFlag{Name: "proto", Aliases: []string{"p"}},
					&cli.StringFlag{Name: "sport", Aliases: []string{"dp"}},
					&cli.StringFlag{Name: "dport", Aliases: []string{"sp"}},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "remove a new acl rule",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "src", Aliases: []string{"s"}},
					&cli.StringFlag{Name: "dst", Aliases: []string{"d"}},
					&cli.StringFlag{Name: "proto", Aliases: []string{"p"}},
					&cli.StringFlag{Name: "sport", Aliases: []string{"dp"}},
					&cli.StringFlag{Name: "dport", Aliases: []string{"sp"}},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all acl rules",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	}
}
