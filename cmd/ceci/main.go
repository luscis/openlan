package main

import (
	"flag"
	"os"

	"github.com/luscis/openlan/pkg/access"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func main() {
	mode := "http"
	conf := ""
	flag.StringVar(&mode, "mode", "http", "Proxy mode for http, socks, tcp and name")
	flag.StringVar(&conf, "conf", "ceci.yaml", "The configuration file")
	flag.Parse()

	if !(mode == "http" || mode == "socks" || mode == "tcp" || mode == "name") {
		libol.Warn("Ceci: not support mode:%s", mode)
		os.Exit(1)
	}

	libol.PreNotify()

	if mode == "name" {
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
	} else {
		c := &config.HttpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		p := proxy.NewHttpProxy(c, nil)
		libol.Go(p.Start)
	}

	libol.SdNotify()
	libol.Wait()
}
