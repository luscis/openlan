package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
	"sort"
)

type Policy struct {
	Cmd
}

func (u Policy) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/policy"
	} else {
		return prefix + "/api/policy/" + name
	}
}

func (u Policy) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "name"}} {{ ps -20 "source" }} {{ ps -20 "destination" }}
{{- range . }}
{{ps -16 .Name}} {{ ps -20 .Source }} {{ ps -20 .Dest }}
{{- end }}
`
}

func (u Policy) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.EspPolicy
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	sort.SliceStable(items, func(i, j int) bool {
		ii := items[i]
		jj := items[j]
		return ii.Name+ii.Source > jj.Name+jj.Source
	})
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Policy) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "policy",
		Aliases: []string{"po"},
		Usage:   "IPSec policy configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all xfrm policy",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
