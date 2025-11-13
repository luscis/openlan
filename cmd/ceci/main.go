package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/proxy"
)

func writepid(file string) {
	pid := fmt.Sprintf("%d", os.Getpid())
	if err := os.WriteFile(file, []byte(pid), 0644); err != nil {
		libol.Warn("Ceci: write pid:%s: %s", pid, err)
	}
}

func main() {
	mode := "http"
	conf := ""
	nodate := false
	pidfile := ""

	flag.StringVar(&mode, "mode", "http", "Proxy mode for http, socks, tcp and name")
	flag.StringVar(&conf, "conf", "ceci.yaml", "The configuration file")
	flag.BoolVar(&nodate, "nodate", nodate, "Dont display message datetime")
	flag.StringVar(&pidfile, "write-pid", pidfile, "Write pid to a file")
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
	switch mode {
	case "name":
		c := &config.NameProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewNameProxy(c)
	case "socks":
		c := &config.SocksProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewSocksProxy(c)
	case "tcp":
		c := &config.TcpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewTcpProxy(c)
	default:
		c := &config.HttpProxy{Conf: conf}
		if err := c.Initialize(); err != nil {
			return
		}
		x = proxy.NewHttpProxy(c, nil)
	}

	libol.Go(x.Start)
	if pidfile != "" {
		writepid(pidfile)
	}
	libol.SdNotify()
	libol.Wait()
	x.Stop()
}
