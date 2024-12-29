package main

import (
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func main() {
	c := config.NewHttpProxy()
	if c != nil {
		libol.PreNotify()
		h := proxy.NewHttpProxy(c, nil)
		libol.SdNotify()
		libol.Go(h.Start)
		libol.Wait()
	}
}
