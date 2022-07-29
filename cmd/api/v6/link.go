package v6

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/database"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/urfave/cli/v2"
	"sort"
	"strings"
)

type Link struct {
}

func (l Link) List(c *cli.Context) error {
	var lsLn []database.VirtualLink
	network := c.String("network")
	if err := database.Client.WhereList(
		func(l *database.VirtualLink) bool {
			return network == "" || l.Network == network
		}, &lsLn); err != nil {
		return err
	} else {
		sort.SliceStable(lsLn, func(i, j int) bool {
			ii := lsLn[i]
			jj := lsLn[j]
			return ii.Network+ii.UUID > jj.Network+jj.UUID
		})
		return api.Out(lsLn, c.String("format"), "")
	}
}

func GetUserPassword(auth string) (string, string) {
	values := strings.SplitN(auth, ":", 2)
	if len(values) == 2 {
		return values[0], values[1]
	}
	return auth, auth
}

func GetDeviceName(conn, device string) string {
	if libol.GetPrefix(conn, 4) == "spi:" {
		return conn
	} else {
		return device
	}
}

func (l Link) Add(c *cli.Context) error {
	auth := c.String("authentication")
	connection := c.String("connection")
	device := c.String("device")
	lsLn := database.VirtualLink{
		UUID:       c.String("uuid"),
		Network:    c.String("network"),
		Connection: connection,
		Device:     device,
	}
	remoteAddr := c.String("remote-address")
	user, pass := GetUserPassword(auth)
	if err := database.Client.Get(&lsLn); err == nil {
		lsVn := database.VirtualNetwork{
			Name: lsLn.Network,
		}
		if lsVn.Name == "" {
			return libol.NewErr("network is nil")
		}
		if err := database.Client.Get(&lsVn); err != nil {
			return libol.NewErr("find network %s: %s", lsVn.Name, err)
		}
		newLn := lsLn
		if connection != "" {
			newLn.Connection = connection
		}
		if user != "" {
			newLn.Authentication["username"] = user
		}
		if pass != "" {
			newLn.Authentication["password"] = pass
		}
		if remoteAddr != "" {
			newLn.OtherConfig["remote_address"] = remoteAddr
		}
		if device != "" {
			newLn.Device = device
		}
		ops, err := database.Client.Where(&lsLn).Update(&newLn)
		if err != nil {
			return err
		}
		if ret, err := database.Client.Transact(ops...); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	} else {
		lsVn := database.VirtualNetwork{
			Name: c.String("network"),
		}
		if lsVn.Name == "" {
			return libol.NewErr("network is nil")
		}
		if err := database.Client.Get(&lsVn); err != nil {
			return libol.NewErr("find network %s: %s", lsVn.Name, err)
		}
		uuid := c.String("uuid")
		if uuid == "" {
			uuid = database.GenUUID()
		}
		newLn := database.VirtualLink{
			Network:    lsLn.Network,
			Connection: lsLn.Connection,
			UUID:       uuid,
			Device:     GetDeviceName(connection, device),
			Authentication: map[string]string{
				"username": user,
				"password": pass,
			},
			OtherConfig: map[string]string{
				"local_address":  lsVn.Address,
				"remote_address": remoteAddr,
			},
		}
		ops, err := database.Client.Create(&newLn)
		if err != nil {
			return err
		}
		libol.Debug("Link.Add %s %s", ops, lsVn)
		database.Client.Execute(ops)
		ops, err = database.Client.Where(&lsVn).Mutate(&lsVn, model.Mutation{
			Field:   &lsVn.LocalLinks,
			Mutator: ovsdb.MutateOperationInsert,
			Value:   []string{newLn.UUID},
		})
		if err != nil {
			return err
		}
		libol.Debug("Link.Add %s", ops)
		database.Client.Execute(ops)
		if ret, err := database.Client.Commit(); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	}
	return nil
}

func (l Link) Remove(c *cli.Context) error {
	lsLn := database.VirtualLink{
		Network:    c.String("network"),
		Connection: c.String("connection"),
		UUID:       c.String("uuid"),
	}
	if err := database.Client.Get(&lsLn); err != nil {
		return err
	}
	lsVn := database.VirtualNetwork{
		Name: lsLn.Network,
	}
	if err := database.Client.Get(&lsVn); err != nil {
		return libol.NewErr("find network %s: %s", lsVn.Name, err)
	}
	if err := database.Client.Get(&lsLn); err != nil {
		return err
	}
	ops, err := database.Client.Where(&lsLn).Delete()
	if err != nil {
		return err
	}
	libol.Debug("Link.Remove %s", ops)
	database.Client.Execute(ops)
	ops, err = database.Client.Where(&lsVn).Mutate(&lsVn, model.Mutation{
		Field:   &lsVn.LocalLinks,
		Mutator: ovsdb.MutateOperationDelete,
		Value:   []string{lsLn.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Link.Remove %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (l Link) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "link",
		Aliases: []string{"li"},
		Usage:   "Virtual Link",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List virtual links",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
				},
				Action: l.List,
			},
			{
				Name:  "add",
				Usage: "Add a virtual link",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "uuid",
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
					&cli.StringFlag{
						Name:  "connection",
						Value: "any",
						Usage: "connection for remote server",
					},
					&cli.StringFlag{
						Name:  "device",
						Usage: "the device name, like spi:10",
					},
					&cli.StringFlag{
						Name:  "authentication",
						Usage: "user and password for authentication",
					},
					&cli.StringFlag{
						Name:  "remote-address",
						Usage: "remote address in this link",
					},
				},
				Action: l.Add,
			},
			{
				Name:  "del",
				Usage: "Del a virtual link",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "uuid",
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
					&cli.StringFlag{
						Name:  "connection",
						Usage: "connection for remote server",
					},
				},
				Action: l.Remove,
			},
		},
	})
}
