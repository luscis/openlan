package main

import (
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cswitch "github.com/luscis/openlan/pkg/switch"
)

func main() {
	c := config.NewSwitch()
	config.Update(c)

	libol.SetLogger(c.Log.File, c.Log.Verbose)
	cache.Init(&c.Perf)
	s := cswitch.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
