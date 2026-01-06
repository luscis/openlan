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
	User{}.Commands(app)
	ACL{}.Commands(app)
	Ldap{}.Commands(app)
	Device{}.Commands(app)
	Lease{}.Commands(app)
	Config{}.Commands(app)
	Server{}.Commands(app)
	Network{}.Commands(app)
	PProf{}.Commands(app)
	IPSec{}.Commands(app)
	Version{}.Commands(app)
	Log{}.Commands(app)
	ZTrust{}.Commands(app)
	RateLimit{}.Commands(app)
	BGP{}.Commands(app)
	Ceci{}.Commands(app)
	Index{}.Commands(app)
	Router{}.Commands(app)
}
