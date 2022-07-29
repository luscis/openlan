// +build linux

package main

import (
	"github.com/luscis/openlan/pkg/access"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

func main() {
	c := config.NewPoint()
	p := access.NewPoint(c)
	// terminal off for linux service, on for open a terminal
	// and others just wait.
	if c.Terminal == "off" {
		libol.PreNotify()
	}
	p.Initialize()
	libol.Go(p.Start)
	if c.Terminal == "on" {
		t := access.NewTerminal(p)
		t.Start()
	} else if c.Terminal == "off" {
		libol.SdNotify()
		libol.Wait()
	} else {
		libol.Wait()
	}
	p.Stop()
}
