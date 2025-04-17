package main

import (
	"flag"
	"os"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func main() {
	mode := "http"
	conf := ""
	nodate := false

	flag.StringVar(&mode, "mode", "http", "Proxy mode for http, socks, tcp and name")
	flag.StringVar(&conf, "conf", "ceci.yaml", "The configuration file")
	flag.BoolVar(&nodate, "nodate", nodate, "Dont display message datetime")
	flag.Parse()

	if nodate {
		libol.NoLogDate()
	}

	if !(mode == "http" || mode == "socks" || mode == "tcp" || mode == "name") {
		libol.Warn("Ceci: not support mode:%s", mode)
		os.Exit(1)
	}

	libol.PreNotify()

	var x proxy.Proxyer
	if mode == "name" {
		c := &config.NameProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewNameProxy(c)
	} else if mode == "socks" {
		c := &config.SocksProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewSocksProxy(c)
	} else if mode == "tcp" {
		c := &config.TcpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewTcpProxy(c)
	} else {
		c := &config.HttpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewHttpProxy(c, nil)
	}

	libol.Go(x.Start)
	libol.SdNotify()
	libol.Wait()
	x.Stop()
}
