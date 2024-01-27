package cswitch

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

const (
	DhcpBin = "dnsmasq"
	DhcpDir = "/var/openlan/dhcp"
)

type Dhcp struct {
	cfg  *co.Dhcp
	out  *libol.SubLogger
	uuid string
}

func NewDhcp(cfg *co.Dhcp) *Dhcp {
	return &Dhcp{
		uuid: cfg.Name,
		cfg:  cfg,
		out:  libol.NewSubLogger(cfg.Name),
	}
}

func (d *Dhcp) Initialize() {
}

func (d *Dhcp) Conf() *co.Dhcp {
	return d.cfg
}

func (d *Dhcp) UUID() string {
	return d.uuid
}

func (d *Dhcp) Path() string {
	return DhcpBin
}

func (d *Dhcp) PidFile() string {
	return filepath.Join(DhcpDir, d.uuid+".pid")
}

func (d *Dhcp) LeaseFile() string {
	return filepath.Join(DhcpDir, d.uuid+".leases")
}

func (d *Dhcp) ConfFile() string {
	return filepath.Join(DhcpDir, d.uuid+".conf")
}

func (d *Dhcp) LogFile() string {
	return filepath.Join(DhcpDir, d.uuid+".log")
}

func (d *Dhcp) Tmpl() string {
	return `#Generate by OpenLAN
strict-order
except-interface=lo
bind-interfaces
interface=%s
dhcp-range=%s,%s,12h
dhcp-leasefile=%s
## disable default gateway
# dhcp-option=3
## disable dns
# dhcp-option=6
`
}

func (d *Dhcp) SaveConf() {
	cfg := d.cfg
	data := fmt.Sprintf(d.Tmpl(),
		cfg.Interface,
		cfg.Subnet.Start,
		cfg.Subnet.End,
		d.LeaseFile(),
	)
	_ = ioutil.WriteFile(d.ConfFile(), []byte(data), 0600)
}

func (d *Dhcp) Start() {
	log, err := libol.CreateFile(d.LogFile())
	if err != nil {
		d.out.Warn("Dhcp.Start %s", err)
		return
	}
	d.SaveConf()
	libol.Go(func() {
		args := []string{
			"--conf-file=" + d.ConfFile(),
			"--pid-file=" + d.PidFile(),
		}
		d.out.Info("Dhcp.Start %s %v", d.Path(), args)
		cmd := exec.Command(d.Path(), args...)
		cmd.Stdout = log
		cmd.Stderr = log
		if err := cmd.Run(); err != nil {
			d.out.Error("Dhcp.Start %s: %s", d.uuid, err)
		}
	})
}

func (d *Dhcp) Clean() {
	files := []string{
		d.LogFile(), d.PidFile(), d.ConfFile(),
	}
	for _, file := range files {
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				d.out.Warn("Dhcp.Clean %s", err)
			}
		}
	}
}

func (d *Dhcp) Stop() {
	if data, err := ioutil.ReadFile(d.PidFile()); err != nil {
		d.out.Info("Dhcp.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			d.out.Warn("Dhcp.Stop %s: %s", pid, err)
		}
	}
	d.Clean()
}
