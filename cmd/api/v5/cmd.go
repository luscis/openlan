package v5

import (
	"os"
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/urfave/cli/v2"
)

func Before(c *cli.Context) error {
	token := c.String("token")
	if token == "" {
		tokenFile := api.AdminTokenFile
		if data, err := os.ReadFile(tokenFile); err == nil {
			token = strings.TrimSpace(string(data))
		}
		_ = c.Set("token", token)
	}
	return nil
}

func After(c *cli.Context) error {
	return nil
}

func Commands(app *api.App) {
	app.After = After
	app.Before = Before
	Version{}.Commands(app)
	User{}.Commands(app)
	Network{}.Commands(app)
	Lease{}.Commands(app)
	ACL{}.Commands(app)
	Ldap{}.Commands(app)
	PProf{}.Commands(app)
	Device{}.Commands(app)
	Config{}.Commands(app)
	Server{}.Commands(app)

	Log{}.Commands(app)
	ZTrust{}.Commands(app)
	RateLimit{}.Commands(app)

	Index{}.Commands(app)
	BGP{}.Commands(app)
	Ceci{}.Commands(app)
	IPSec{}.Commands(app)
	Router{}.Commands(app)
	Reload{}.Commands(app)
}
