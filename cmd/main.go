package main

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/cmd/api/v5"
	"github.com/luscis/openlan/cmd/api/v6"
	"log"
	"os"
)

func main() {
	api.Version = api.GetEnv("VERSION", api.Version)
	api.Url = api.GetEnv("URL", api.Url)
	api.Token = api.GetEnv("TOKEN", api.Token)
	api.Server = api.GetEnv("CONFSERVER", api.Server)
	api.Database = api.GetEnv("DATABASE", api.Database)
	app := &api.App{}
	app.New()

	switch api.Version {
	case "v6":
		v6.Commands(app)
	default:
		v5.Commands(app)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
