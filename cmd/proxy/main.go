package main

import (
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func main() {
	c := config.NewProxy()
	libol.SetLogger(c.Log.File, c.Log.Verbose)

	p := proxy.NewProxy(c)
	libol.PreNotify()
	p.Initialize()
	if c.PProf != "" {
		f := libol.PProf{Listen: c.PProf}
		f.Start()
	}
	libol.Go(p.Start)
	libol.SdNotify()
	libol.Wait()
	p.Stop()
}
