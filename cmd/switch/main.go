package main

import (
	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/switch"
	"log"
)

func main() {
	log.SetFlags(0)
	udp := api.GetEnv("ESPUDP", "4500")
	config.SetLocalUdp(udp)
	c := config.NewSwitch()
	libol.SetLogger(c.Log.File, c.Log.Verbose)
	libol.Debug("main %s", c)
	cache.Init(&c.Perf)
	s := _switch.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
