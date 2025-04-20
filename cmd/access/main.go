package main

import (
	"github.com/luscis/openlan/pkg/access"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

func main() {
	c := config.NewPoint()
	p := access.NewPoint(c)

	p.Initialize()
	libol.Go(p.Start)

	libol.Wait()
	p.Stop()
}
