package main

import (
	"flag"

	"github.com/luscis/openlan/pkg/access"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func main() {
	mode := "http"
	conf := ""
	flag.StringVar(&mode, "mode", "http", "Proxy mode for http, socks, tcp and access")
	flag.StringVar(&conf, "conf", "ceci.yaml", "The configuration file")
	flag.Parse()

	libol.PreNotify()
	if mode == "http" {
		c := &config.HttpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		p := proxy.NewHttpProxy(c, nil)
		libol.Go(p.Start)
	} else if mode == "socks" {
		c := &config.SocksProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		p := proxy.NewSocksProxy(c)
		libol.Go(p.Start)
	} else if mode == "tcp" {
		c := &config.TcpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		p := proxy.NewTcpProxy(c)
		libol.Go(p.Start)
	} else if mode == "access" {
		c := &config.Point{
			RequestAddr: true,
			Terminal:    "off",
			Conf:        conf,
		}
		if err := c.Initialize(); err != nil {
			return
		}
		p := access.NewPoint(c)
		p.Initialize()
		libol.Go(p.Start)
	}
	libol.SdNotify()
	libol.Wait()
}
