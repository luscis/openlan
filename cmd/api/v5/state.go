package v5

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/urfave/cli/v2"
	"sort"
)

type State struct {
	Cmd
}

func (u State) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/state"
	} else {
		return prefix + "/api/state/" + name
	}
}

func (u State) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "name"}} {{ps -8 "spi"}} {{ ps -16 "local" }} {{ ps -16 "remote" }} {{ ps -12 "rx bytes" }} {{ ps -12 "tx bytes" }} {{ ps -12 "rx packages" }} {{ ps -12 "tx packages" }}
{{- range . }}
{{ps -16 .Name}} {{pi -8 .Spi }} {{ ps -16 .Local }} {{ ps -16 .Remote }} {{ pi -12 .RxBytes }} {{ pi -12 .TxBytes }} {{ pi -12 .RxPackages }} {{ pi -12 .TxPackages }}
{{- end }}
`
}

func (u State) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.EspState
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	sort.SliceStable(items, func(i, j int) bool {
		ii := items[i]
		jj := items[j]
		return ii.Spi > jj.Spi
	})
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u State) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "state",
		Aliases: []string{"se"},
		Usage:   "IPSec state configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all xfrm state",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
