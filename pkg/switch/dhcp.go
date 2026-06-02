package cswitch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
%s%s
`
}

func (d *Dhcp) SaveConf() {
	cfg := d.cfg
	gateway := "## disable default gateway\n# dhcp-option=3\n"
	if cfg.Gateway != "" {
		gateway = fmt.Sprintf("dhcp-option=3,%s\n", cfg.Gateway)
	}
	dns := "## disable dns\n# dhcp-option=6\n"
	if len(cfg.DNS) > 0 {
		dns = fmt.Sprintf("dhcp-option=6,%s\n", strings.Join(cfg.DNS, ","))
	}
	data := fmt.Sprintf(d.Tmpl(),
		cfg.Interface,
		cfg.Subnet.Start,
		cfg.Subnet.End,
		d.LeaseFile(),
		gateway,
		dns,
	)
	_ = os.WriteFile(d.ConfFile(), []byte(data), 0600)
}

func (d *Dhcp) Start() {
	d.SaveConf()
	libol.Go(func() {
		args := []string{
			"--conf-file=" + d.ConfFile(),
			"--pid-file=" + d.PidFile(),
			"--log-facility=" + d.LogFile(),
		}
		d.out.Info("Dhcp.Start %s %v", d.Path(), args)
		cmd := exec.Command(d.Path(), args...)
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
	if data, err := os.ReadFile(d.PidFile()); err != nil {
		d.out.Info("Dhcp.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		if pidNum, err := strconv.Atoi(pid); err != nil {
			d.out.Warn("Dhcp.Stop %s: %s", pid, err)
		} else if _, err := libol.Kill(pidNum); err != nil {
			d.out.Warn("Dhcp.Stop %s: %s", pid, err)
		} else {
			for i := 0; i < 20; i++ {
				if !libol.HasProcess(pidNum) {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	d.Clean()
}
