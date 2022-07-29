package v6

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/database"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/urfave/cli/v2"
	"sort"
)

type Prefix struct {
}

func (u Prefix) List(c *cli.Context) error {
	var list []database.PrefixRoute
	if err := database.Client.List(&list); err != nil {
		return err
	} else {
		sort.SliceStable(list, func(i, j int) bool {
			ii := list[i]
			jj := list[j]
			return ii.UUID > jj.UUID
		})
		return api.Out(list, c.String("format"), "")
	}
}

func (u Prefix) Add(c *cli.Context) error {
	lsVn := database.VirtualNetwork{
		Name: c.String("network"),
	}
	if lsVn.Name == "" {
		return libol.NewErr("network is nil")
	}
	if err := database.Client.Get(&lsVn); err != nil {
		return libol.NewErr("find network %s: %s", lsVn.Name, err)
	}
	newPf := database.PrefixRoute{
		UUID:    database.GenUUID(),
		Network: lsVn.Name,
		Source:  c.String("source"),
		Prefix:  c.String("prefix"),
		Gateway: c.String("gateway"),
		Mode:    c.String("mode"),
	}
	ops, err := database.Client.Create(&newPf)
	if err != nil {
		return err
	}
	libol.Debug("Prefix.Add %s %s", ops, lsVn)
	database.Client.Execute(ops)
	ops, err = database.Client.Where(&lsVn).Mutate(&lsVn, model.Mutation{
		Field:   &lsVn.PrefixRoutes,
		Mutator: ovsdb.MutateOperationInsert,
		Value:   []string{newPf.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Prefix.Add %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (u Prefix) Remove(c *cli.Context) error {
	lsPf := database.PrefixRoute{
		Network: c.String("network"),
		Prefix:  c.String("prefix"),
		UUID:    c.String("uuid"),
	}
	if err := database.Client.Get(&lsPf); err != nil {
		return err
	}
	lsVn := database.VirtualNetwork{
		Name: lsPf.Network,
	}
	if err := database.Client.Get(&lsVn); err != nil {
		return libol.NewErr("find network %s: %s", lsVn.Name, err)
	}
	if err := database.Client.Get(&lsPf); err != nil {
		return err
	}
	ops, err := database.Client.Where(&lsPf).Delete()
	if err != nil {
		return err
	}
	libol.Debug("Prefix.Remove %s", ops)
	database.Client.Execute(ops)
	ops, err = database.Client.Where(&lsVn).Mutate(&lsVn, model.Mutation{
		Field:   &lsVn.PrefixRoutes,
		Mutator: ovsdb.MutateOperationDelete,
		Value:   []string{lsPf.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Prefix.Remove %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (u Prefix) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "route",
		Aliases: []string{"ro"},
		Usage:   "Prefix route",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List prefix routes",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add a prefix route",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
					&cli.StringFlag{
						Name: "prefix",
					},
					&cli.StringFlag{
						Name:  "source",
						Value: "0.0.0.0/0",
					},
					&cli.StringFlag{
						Name:  "gateway",
						Value: "local",
					},
					&cli.StringFlag{
						Name:  "mode",
						Value: "direct",
					},
				},
				Action: u.Add,
			},
			{
				Name:  "del",
				Usage: "delete a prefix route",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "uuid",
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
					&cli.StringFlag{
						Name: "prefix",
					},
				},
				Action: u.Remove,
			},
		},
	})
}
