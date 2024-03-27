package cswitch

import (
	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/database"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/model"
)

type ConfD struct {
	stop chan struct{}
	out  *libol.SubLogger
	api  api.Switcher
}

func NewConfd(api api.Switcher) *ConfD {
	c := &ConfD{
		out:  libol.NewSubLogger("confd"),
		stop: make(chan struct{}),
		api:  api,
	}
	return c
}

func (c *ConfD) Initialize() {
}

func (c *ConfD) Start() {
	c.out.Info("ConfD.Start")
	handler := &cache.EventHandlerFuncs{
		AddFunc:    c.Add,
		DeleteFunc: c.Delete,
		UpdateFunc: c.Update,
	}
	if _, err := database.NewConfClient(handler); err != nil {
		c.out.Error("ConfD.Start open db with %s", err)
		return
	}
}

func (c *ConfD) Stop() {
	c.out.Info("ConfD.Stop")
}

func (c *ConfD) Add(table string, model model.Model) {
	c.out.Cmd("ConfD.Add %s %v", table, model)
	if obj, ok := model.(*database.Switch); ok {
		c.out.Info("ConfD.Add switch %d", obj.Listen)
	}

	if obj, ok := model.(*database.VirtualNetwork); ok {
		c.out.Info("ConfD.Add virtual network %s:%s", obj.Name, obj.Address)
	}

	if obj, ok := model.(*database.VirtualLink); ok {
		c.out.Info("ConfD.Add virtual link %s:%s", obj.Network, obj.Connection)
		c.AddLink(obj)
	}

	if obj, ok := model.(*database.NameCache); ok {
		c.out.Info("ConfD.Add name cache %s", obj.Name)
		c.UpdateName(obj)
	}

	if obj, ok := model.(*database.PrefixRoute); ok {
		c.out.Info("ConfD.Add prefix route %s:%s", obj.Network, obj.Prefix)
		c.AddRoute(obj)
	}
}

func (c *ConfD) Delete(table string, model model.Model) {
	c.out.Cmd("ConfD.Delete %s %v", table, model)
	if obj, ok := model.(*database.VirtualNetwork); ok {
		c.out.Info("ConfD.Delete virtual network %s:%s", obj.Name, obj.Address)
	}

	if obj, ok := model.(*database.VirtualLink); ok {
		c.out.Info("ConfD.Delete virtual link %s:%s", obj.Network, obj.Connection)
		c.DelLink(obj)
	}

	if obj, ok := model.(*database.PrefixRoute); ok {
		c.out.Info("ConfD.Delete prefix route %s:%s", obj.Network, obj.Prefix)
		c.DelRoute(obj)
	}
}

func (c *ConfD) Update(table string, old model.Model, new model.Model) {
	c.out.Cmd("ConfD.Update %s %v", table, new)
	if obj, ok := new.(*database.VirtualNetwork); ok {
		c.out.Info("ConfD.Update virtual network %s:%s", obj.Name, obj.Address)
	}

	if ob1, ok := new.(*database.VirtualLink); ok {
		ob0, _ := old.(*database.VirtualLink)
		oldConn := ob0.Status["remote_connection"]
		newConn := ob1.Status["remote_connection"]
		c.out.Info("ConfD.Update virtual link %s:%s->%s", ob1.Network, oldConn, newConn)
		if c.DiffLink(ob0, ob1) {
			c.AddLink(ob1)
		}
	}

	if obj, ok := new.(*database.NameCache); ok {
		c.out.Info("ConfD.Update name cache %s", obj.Name)
		c.UpdateName(obj)
	}
}

func GetRoutes(result *[]database.PrefixRoute, device string) error {
	if err := database.Client.WhereList(
		func(l *database.PrefixRoute) bool {
			return l.Gateway == device
		}, result); err != nil {
		libol.Warn("GetRoutes %v has %s", device, err)
		return err
	}
	return nil
}

func (c *ConfD) DiffLink(old *database.VirtualLink, new *database.VirtualLink) bool {
	c.out.Cmd("ConfD.Update virtual link %v->%v", old, new)
	if old.Connection != new.Connection {
		return true
	}
	if old.Device != new.Device {
		return true
	}
	if old.Status["remote_connection"] != new.Status["remote_connection"] {
		return true
	}
	if old.OtherConfig["local_address"] != new.OtherConfig["local_address"] {
		return true
	}
	if old.OtherConfig["remote_address"] != new.OtherConfig["remote_address"] {
		return true
	}
	if old.Authentication["username"] != new.Authentication["username"] {
		return true
	}
	if old.Authentication["password"] != new.Authentication["password"] {
		return true
	}
	return false
}

func (c *ConfD) AddLink(obj *database.VirtualLink) {
	worker := api.GetWorker(obj.Network)
	if worker == nil {
		c.out.Warn("ConfD.AddLink network %s not found.", obj.Network)
		return
	}
	cfg := worker.Config()
	if cfg == nil || cfg.Specifies == nil {
		c.out.Warn("ConfD.AddLink config %s not found.", obj.Network)
		return
	}
	if cfg.Provider == "esp" {
		link := &MemberLink{
			LinkImpl{
				api:    c.api,
				out:    c.out,
				worker: worker,
			},
		}
		link.Add(obj)
	} else if cfg.Provider == "fabric" {
		link := &TunnelLink{
			LinkImpl{
				api:    c.api,
				out:    c.out,
				worker: worker,
			},
		}
		link.Add(obj)
	}
}

func (c *ConfD) DelLink(obj *database.VirtualLink) {
	worker := api.GetWorker(obj.Network)
	if worker == nil {
		c.out.Warn("ConfD.DelLink network %s not found.", obj.Network)
		return
	}
	cfg := worker.Config()
	if cfg == nil || cfg.Specifies == nil {
		c.out.Warn("ConfD.DelLink config %s not found.", obj.Network)
		return
	}
	if cfg.Provider == "esp" {
		link := &MemberLink{
			LinkImpl{
				api:    c.api,
				out:    c.out,
				worker: worker,
			},
		}
		link.Del(obj)
	} else if cfg.Provider == "fabric" {
		link := &TunnelLink{
			LinkImpl{
				api:    c.api,
				out:    c.out,
				worker: worker,
			},
		}
		link.Del(obj)
	}
}

func (c *ConfD) UpdateName(obj *database.NameCache) {
	if obj.Address == "" {
		return
	}
	c.out.Info("ConfD.UpdateName %s %s", obj.Name, obj.Address)
	api.ListWorker(func(w api.Networker) {
		cfg := w.Config()
		spec := cfg.Specifies
		if spec == nil {
			return
		}
		if specObj, ok := spec.(*config.EspSpecifies); ok {
			if specObj.HasRemote(obj.Name, obj.Address) {
				cs := config.Get()
				cfg.Correct(cs)
				w.Reload(c.api)
			}
		}
	})
}

func (c *ConfD) AddRoute(obj *database.PrefixRoute) {
	if obj.Prefix == "" {
		return
	}
	c.out.Cmd("ConfD.DelRoute %v", obj.Network)
	worker := api.GetWorker(obj.Network)
	if worker == nil {
		c.out.Warn("ConfD.DelRoute network %s not found.", obj.Network)
		return
	}
	netCfg := worker.Config()
	if netCfg == nil || netCfg.Specifies == nil {
		c.out.Warn("ConfD.DelRoute config %s not found.", obj.Network)
		return
	}
	spec := netCfg.Specifies
	poCfg := &config.EspPolicy{
		Source: obj.Source,
		Dest:   obj.Prefix,
	}
	if specObj, ok := spec.(*config.EspSpecifies); ok {
		var mem *config.EspMember
		if mem = specObj.GetMember(obj.Gateway); mem != nil {
			mem.AddPolicy(poCfg)
		} else if libol.GetPrefix(obj.Gateway, 4) == "spi:" {
			mem = &config.EspMember{
				Name: obj.Gateway,
			}
			specObj.AddMember(mem)
		}
		if mem != nil {
			mem.AddPolicy(poCfg)
			specObj.Correct()
			worker.Reload(c.api)
		}
	}
}

func (c *ConfD) DelRoute(obj *database.PrefixRoute) {
	if obj.Prefix == "" {
		return
	}
	c.out.Cmd("ConfD.DelRoute %v", obj.Network)
	worker := api.GetWorker(obj.Network)
	if worker == nil {
		c.out.Warn("ConfD.DelRoute network %s not found.", obj.Network)
		return
	}
	netCfg := worker.Config()
	if netCfg == nil || netCfg.Specifies == nil {
		c.out.Warn("ConfD.DelRoute config %s not found.", obj.Network)
		return
	}
	spec := netCfg.Specifies
	if specObj, ok := spec.(*config.EspSpecifies); ok {
		if mem := specObj.GetMember(obj.Gateway); mem != nil {
			if mem.RemovePolicy(obj.Prefix) {
				specObj.Correct()
				worker.Reload(c.api)
			}
		}
	}
}

type LinkImpl struct {
	api    api.Switcher
	out    *libol.SubLogger
	worker api.Networker
}

func (l *LinkImpl) Add(obj *database.VirtualLink) {
	l.out.Info("LinkImpl.Add TODO")
}

func (l *LinkImpl) Update(obj *database.VirtualLink) {
	l.out.Info("LinkImpl.Update TODO")
}

func (l *LinkImpl) Del(obj *database.VirtualLink) {
	l.out.Info("LinkImpl.Del TODO")
}

type MemberLink struct {
	LinkImpl
}

func (l *MemberLink) Add(obj *database.VirtualLink) {
	var port int
	var remote string

	conn := obj.Connection
	if conn == "any" {
		remoteConn := obj.Status["remote_connection"]
		if libol.GetPrefix(remoteConn, 4) == "udp:" {
			remote, port = database.GetAddrPort(remoteConn[4:])
		} else {
			l.out.Warn("MemberLink.Add %s remote not found.", conn)
			return
		}
	} else if libol.GetPrefix(conn, 4) == "udp:" {
		remoteConn := obj.Connection
		remote, port = database.GetAddrPort(remoteConn[4:])
	} else {
		return
	}
	l.out.Info("MemberLink.Add remote link %s:%d", remote, port)
	memCfg := &config.EspMember{
		Name:    obj.Device,
		Address: obj.OtherConfig["local_address"],
		Peer:    obj.OtherConfig["remote_address"],
		State: config.EspState{
			Remote:     remote,
			RemotePort: port,
			Auth:       obj.Authentication["password"],
			Crypt:      obj.Authentication["username"],
		},
	}
	var routes []database.PrefixRoute
	_ = GetRoutes(&routes, obj.Device)
	for _, route := range routes {
		l.out.Info("MemberLink.Add %s via %s", route.Prefix, obj.Device)
		memCfg.AddPolicy(&config.EspPolicy{
			Source: route.Source,
			Dest:   route.Prefix,
		})
	}
	l.out.Cmd("MemberLink.Add %v", memCfg)
	spec := l.worker.Config().Specifies
	if specObj, ok := spec.(*config.EspSpecifies); ok {
		specObj.AddMember(memCfg)
		specObj.Correct()
		l.worker.Reload(l.api)
	}
}

func (l *MemberLink) Update(obj *database.VirtualLink) {

}

func (l *MemberLink) Del(obj *database.VirtualLink) {
	l.out.Info("MemberLink.Del remote link %s", obj.Device)
	spec := l.worker.Config().Specifies
	if specObj, ok := spec.(*config.EspSpecifies); ok {
		if specObj.DelMember(obj.Device) {
			specObj.Correct()
			l.worker.Reload(l.api)
		}
	}
}

type TunnelLink struct {
	LinkImpl
}

func (l *TunnelLink) Add(obj *database.VirtualLink) {
	tunCfg := &config.FabricTunnel{
		Remote: obj.Connection,
	}
	l.out.Cmd("TunnelLink.Add %v", tunCfg)
	spec := l.worker.Config().Specifies
	if specObj, ok := spec.(*config.FabricSpecifies); ok {
		specObj.AddTunnel(tunCfg)
		specObj.Correct()
		l.worker.Reload(l.api)
	}
}

func (l *TunnelLink) Del(obj *database.VirtualLink) {
	l.out.Info("TunnelLink.Del remote link %s", obj.Connection)
	spec := l.worker.Config().Specifies
	if specObj, ok := spec.(*config.FabricSpecifies); ok {
		if specObj.DelTunnel(obj.Connection) {
			specObj.Correct()
			l.worker.Reload(l.api)
		}
	}
}
