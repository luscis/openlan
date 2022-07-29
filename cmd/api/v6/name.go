package v6

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/database"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
	"net"
	"sort"
	"time"
)

type Name struct {
}

func (u Name) List(c *cli.Context) error {
	var listNa []database.NameCache
	if err := database.Client.List(&listNa); err != nil {
		return err
	} else {
		sort.SliceStable(listNa, func(i, j int) bool {
			ii := listNa[i]
			jj := listNa[j]
			return ii.UUID > jj.UUID
		})
		return api.Out(listNa, c.String("format"), "")
	}
}

func (u Name) Add(c *cli.Context) error {
	name := c.String("name")
	lsNa := database.NameCache{
		Name: name,
		UUID: c.String("uuid"),
	}
	if lsNa.Name == "" && lsNa.UUID == "" {
		return libol.NewErr("Name is nil")
	}
	address := c.String("address")
	if address == "" {
		addrIps, _ := net.LookupIP(lsNa.Name)
		if len(addrIps) > 0 {
			address = addrIps[0].String()
		}
	}
	newNa := lsNa
	if name != "" {
		newNa.Name = name
	}
	if address != "" {
		newNa.Address = address
	}
	newNa.UpdateAt = time.Now().Format("2006-01-02T15:04")
	if err := database.Client.Get(&lsNa); err == nil {
		if lsNa.Address != address {
			ops, err := database.Client.Where(&lsNa).Update(&newNa)
			if err != nil {
				return err
			}
			if ret, err := database.Client.Transact(ops...); err != nil {
				return err
			} else {
				database.PrintError(ret)
			}
		}
	} else {
		ops, err := database.Client.Create(&newNa)
		if err != nil {
			return err
		}
		libol.Debug("Name.Add %s", ops)
		if ret, err := database.Client.Transact(ops...); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	}
	return nil
}

func (u Name) Remove(c *cli.Context) error {
	lsNa := database.NameCache{
		Name: c.String("name"),
		UUID: c.String("uuid"),
	}
	if err := database.Client.Get(&lsNa); err != nil {
		return nil
	}
	ops, err := database.Client.Where(&lsNa).Delete()
	if err != nil {
		return err
	}
	libol.Debug("Name.Remove %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (u Name) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "name",
		Aliases: []string{"na"},
		Usage:   "Name cache",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List name cache",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add or update name cache",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "uuid",
					},
					&cli.StringFlag{
						Name: "name",
					},
					&cli.StringFlag{
						Name: "address",
					},
				},
				Action: u.Add,
			},
			{
				Name:  "del",
				Usage: "Delete a name cache",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "uuid",
					},
					&cli.StringFlag{
						Name: "name",
					},
				},
				Action: u.Remove,
			},
		},
	})
}
