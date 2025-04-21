package v5

import (
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
)

type Access struct {
	Cmd
}

func (u Access) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/point"
	} else {
		return prefix + "/api/point/" + name
	}
}

func (u Access) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "uuid"}} {{ps -8 "alive"}} {{ ps -8 "device" }} {{ps -16 "alias"}} {{ps -8 "user"}} {{ps -22 "remote"}} {{ps -8 "network"}} {{ ps -6 "state"}}
{{- range . }}
{{ps -16 .UUID}} {{pt .AliveTime | ps -8}} {{ ps -8 .Device}} {{ps -16 .Alias}} {{ps -8 .User}} {{ps -22 .Remote}} {{ps -8 .Network}}  {{ ps -6 .State}}
{{- end }}
`
}

func (u Access) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Access
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	name := c.String("name")
	if len(name) > 0 {
		tmp := items[:0]
		for _, obj := range items {
			if obj.Network == name {
				tmp = append(tmp, obj)
			}
		}
		items = tmp
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Access) Commands() *cli.Command {
	return &cli.Command{
		Name:  "access",
		Usage: "access to this switch",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all access",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	}
}
