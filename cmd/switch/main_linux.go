package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cswitch "github.com/luscis/openlan/pkg/switch"
)

func run() {
	c := config.NewSwitch()
	config.Update(c)

	libol.SetLogger(c.Log.File, c.Log.Verbose)
	libol.ShowVersion()

	cache.Init(&c.Limit)
	s := cswitch.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}

const monitor = "-monitor"

func main() {
	var newArgs []string

	exePath := os.Args[0]
	isMonitor := false

	for _, v := range os.Args[1:] {
		if v == monitor {
			isMonitor = true
		} else {
			newArgs = append(newArgs, v)
		}
	}
	if !isMonitor {
		run()
		return
	}

	libol.Info("%s with %s", exePath, newArgs)
	for {
		cmd := exec.Command(exePath, newArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			libol.Error("Exec: %s", err)
			break
		} else {
			cmd.Wait()
		}
		time.Sleep(2 * time.Second)
	}
}
