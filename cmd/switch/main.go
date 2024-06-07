package main

import (
	"log"

	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/switch"
)

func main() {
	log.SetFlags(0)
	c := config.NewSwitch()
	config.Update(c)

	libol.SetLogger(c.Log.File, c.Log.Verbose)
	libol.Debug("main %s", c)
	cache.Init(&c.Perf)
	s := cswitch.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
