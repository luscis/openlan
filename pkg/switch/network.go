package _switch

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/vishvananda/netlink"
)

type Networker interface {
	String() string
	ID() string
	Initialize()
	Start(v api.Switcher)
	Stop()
	Bridge() cn.Bridger
	Config() *co.Network
	Subnet() string
	Reload(v api.Switcher)
	Provider() string
}

var workers = make(map[string]Networker)

func NewNetworker(c *co.Network) Networker {
	var obj Networker
	switch c.Provider {
	case "esp":
		obj = NewESPWorker(c)
	case "vxlan":
		obj = NewVxLANWorker(c)
	case "fabric":
		obj = NewFabricWorker(c)
	default:
		obj = NewOpenLANWorker(c)
	}
	workers[c.Name] = obj
	return obj
}

func GetWorker(name string) Networker {
	return workers[name]
}

func ListWorker(call func(w Networker)) {
	for _, worker := range workers {
		call(worker)
	}
}

type LinuxPort struct {
	name string // gre:xx, vxlan:xx
	vlan int
	link string
}

type WorkerImpl struct {
	uuid    string
	cfg     *co.Network
	out     *libol.SubLogger
	dhcp    *Dhcp
	outputs []*LinuxPort
	fire    *cn.FireWallTable
}

func NewWorkerApi(c *co.Network) *WorkerImpl {
	return &WorkerImpl{
		cfg: c,
		out: libol.NewSubLogger(c.Name),
	}
}

func (w *WorkerImpl) Provider() string {
	return w.cfg.Provider
}

func (w *WorkerImpl) Initialize() {
	if w.cfg.Dhcp == "enable" {
		w.dhcp = NewDhcp(&co.Dhcp{
			Name:   w.cfg.Name,
			Subnet: w.cfg.Subnet,
			Bridge: w.cfg.Bridge,
		})
	}
	w.fire = cn.NewFireWallTable(w.cfg.Name)
}

func (w *WorkerImpl) AddPhysical(bridge string, vlan int, output string) {
	link, err := netlink.LinkByName(output)
	if err != nil {
		w.out.Error("WorkerImpl.LinkByName %s %s", output, err)
		return
	}
	slaver := output
	if vlan > 0 {
		if err := netlink.LinkSetUp(link); err != nil {
			w.out.Warn("WorkerImpl.LinkSetUp %s %s", output, err)
		}
		subLink := &netlink.Vlan{
			LinkAttrs: netlink.LinkAttrs{
				Name:        fmt.Sprintf("%s.%d", output, vlan),
				ParentIndex: link.Attrs().Index,
			},
			VlanId: vlan,
		}
		if err := netlink.LinkAdd(subLink); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", subLink.Name, err)
			return
		}
		slaver = subLink.Name
	}
	br := cn.NewBrCtl(bridge, 0)
	if err := br.AddPort(slaver); err != nil {
		w.out.Warn("WorkerImpl.AddPhysical %s", err)
	}
}

func (w *WorkerImpl) AddOutput(bridge string, port *LinuxPort) {
	name := port.name
	values := strings.SplitN(name, ":", 6)
	if values[0] == "gre" {
		if port.link == "" {
			port.link = co.GenName("ge-")
		}
		link := &netlink.Gretap{
			LinkAttrs: netlink.LinkAttrs{
				Name: port.link,
			},
			Local:    libol.ParseAddr("0.0.0.0"),
			Remote:   libol.ParseAddr(values[1]),
			PMtuDisc: 1,
		}
		if err := netlink.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", name, err)
			return
		}
	} else if values[0] == "vxlan" {
		if len(values) < 3 {
			w.out.Error("WorkerImpl.LinkAdd %s wrong", name)
			return
		}
		if port.link == "" {
			port.link = co.GenName("vn-")
		}
		dport := 8472
		if len(values) == 4 {
			dport, _ = strconv.Atoi(values[3])
		}
		vni, _ := strconv.Atoi(values[2])
		link := &netlink.Vxlan{
			VxlanId: vni,
			LinkAttrs: netlink.LinkAttrs{
				TxQLen: -1,
				Name:   port.link,
			},
			Group: libol.ParseAddr(values[1]),
			Port:  dport,
		}
		if err := netlink.LinkAdd(link); err != nil {
			w.out.Error("WorkerImpl.LinkAdd %s %s", name, err)
			return
		}
	} else {
		port.link = name
	}
	w.out.Info("WorkerImpl.AddOutput %s %s", port.link, port.name)
	w.AddPhysical(bridge, port.vlan, port.link)
}

func (w *WorkerImpl) Start(v api.Switcher) {
	cfg := w.cfg
	fire := w.fire

	if cfg.Acl != "" {
		fire.Raw.Pre.AddRule(cn.IpRule{
			Input: cfg.Bridge.Name,
			Jump:  cfg.Acl,
		})
	}
	fire.Filter.For.AddRule(cn.IpRule{
		Input:  cfg.Bridge.Name,
		Output: cfg.Bridge.Name,
	})
	if cfg.Bridge.Mss > 0 {
		// forward to remote
		fire.Mangle.Post.AddRule(cn.IpRule{
			Output:  cfg.Bridge.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    "TCPMSS",
			SetMss:  cfg.Bridge.Mss,
		})
		// connect from local
		fire.Mangle.In.AddRule(cn.IpRule{
			Input:   cfg.Bridge.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    "TCPMSS",
			SetMss:  cfg.Bridge.Mss,
		})
	}
	for _, output := range cfg.Outputs {
		port := &LinuxPort{
			name: output.Interface,
			vlan: output.Vlan,
		}
		w.AddOutput(cfg.Bridge.Name, port)
		w.outputs = append(w.outputs, port)
	}
	if w.dhcp != nil {
		w.dhcp.Start()
		fire.Nat.Post.AddRule(cn.IpRule{
			Source: cfg.Bridge.Address,
			NoDest: cfg.Bridge.Address,
			Jump:   cn.CMasq,
		})
	}
}

func (w *WorkerImpl) DelPhysical(bridge string, vlan int, output string) {
	if vlan > 0 {
		subLink := &netlink.Vlan{
			LinkAttrs: netlink.LinkAttrs{
				Name: fmt.Sprintf("%s.%d", output, vlan),
			},
		}
		if err := netlink.LinkDel(subLink); err != nil {
			w.out.Error("WorkerImpl.DelPhysical.LinkDel %s %s", subLink.Name, err)
			return
		}
	} else {
		br := cn.NewBrCtl(bridge, 0)
		if err := br.DelPort(output); err != nil {
			w.out.Warn("WorkerImpl.DelPhysical %s", err)
		}
	}
}

func (w *WorkerImpl) DelOutput(bridge string, port *LinuxPort) {
	w.out.Info("WorkerImpl.DelOutput %s %s", port.link, port.name)
	w.DelPhysical(bridge, port.vlan, port.link)
	values := strings.SplitN(port.name, ":", 6)
	if values[0] == "gre" {
		link := &netlink.Gretap{
			LinkAttrs: netlink.LinkAttrs{
				Name: port.link,
			},
		}
		if err := netlink.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.DelOutput.LinkDel %s %s", link.Name, err)
			return
		}
	} else if values[0] == "vxlan" {
		link := &netlink.Vxlan{
			LinkAttrs: netlink.LinkAttrs{
				Name: port.link,
			},
		}
		if err := netlink.LinkDel(link); err != nil {
			w.out.Error("WorkerImpl.DelOutput.LinkDel %s %s", link.Name, err)
			return
		}
	}
}

func (w *WorkerImpl) Stop() {
	if w.dhcp != nil {
		w.dhcp.Stop()
	}
	for _, output := range w.outputs {
		w.DelOutput(w.cfg.Bridge.Name, output)
	}
	w.outputs = nil
}

func (w *WorkerImpl) String() string {
	return w.cfg.Name
}

func (w *WorkerImpl) ID() string {
	return w.uuid
}

func (w *WorkerImpl) Bridge() cn.Bridger {
	return nil
}

func (w *WorkerImpl) Config() *co.Network {
	return w.cfg
}

func (w *WorkerImpl) Subnet() string {
	return ""
}

func (w *WorkerImpl) Reload(v api.Switcher) {
}
