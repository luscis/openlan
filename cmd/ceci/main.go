package main

import (
	"flag"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func main() {
	mode := "http"
	conf := ""
	flag.StringVar(&mode, "mode", "http", "Proxy mode for tcp or http")
	flag.StringVar(&conf, "conf", "ceci.json", "The configuration file")
	flag.Parse()

	libol.PreNotify()
	if mode == "http" {
		c := &config.HttpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		p := proxy.NewHttpProxy(c, nil)
		libol.Go(p.Start)

	} else {
		c := &config.TcpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		p := proxy.NewTcpProxy(c)
		libol.Go(p.Start)

	}
	libol.SdNotify()
	libol.Wait()
}
