package main

import (
	"log"
	"os"

	"github.com/luscis/openlan/cmd/api"
	v5 "github.com/luscis/openlan/cmd/api/v5"
)

func main() {
	api.Url = api.GetEnv("URL", api.Url)
	api.Token = api.GetEnv("TOKEN", api.Token)
	app := &api.App{}
	app.New()

	switch api.Version {
	default:
		v5.Commands(app)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
