package api

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type KernelRoute struct {
}

func (l KernelRoute) Router(router *mux.Router) {
	router.HandleFunc("/api/kernel/route", l.List).Methods("GET")
}

func ListeRoutes() []schema.KernelRoute {
	var items []schema.KernelRoute
	values, _ := libol.ListRoutes()
	for _, val := range values {
		item := schema.KernelRoute{
			Link:     val.Link,
			Metric:   val.Priority,
			Table:    val.Table,
			Source:   val.Src,
			NextHop:  val.Gw,
			Prefix:   val.Dst,
			Protocol: val.Protocol,
		}
		for _, obj := range val.MultiPath {
			item.Multipath = append(item.Multipath,
				schema.MultiPath{
					NextHop: obj.Gw,
					Link:    obj.Link,
				},
			)
		}
		items = append(items, item)
	}
	return items
}
func (l KernelRoute) List(w http.ResponseWriter, r *http.Request) {
	items := ListeRoutes()
	ResponseJson(w, items)
}

type KernelNeighbor struct {
}

func (l KernelNeighbor) Router(router *mux.Router) {
	router.HandleFunc("/api/kernel/neighbor", l.List).Methods("GET")
}

func ListNeighbrs() []schema.KernelNeighbor {
	var items []schema.KernelNeighbor
	values, _ := libol.ListNeighbrs()
	for _, val := range values {
		item := schema.KernelNeighbor{
			Link:    val.Link,
			Address: val.Address,
			HwAddr:  val.HwAddr,
			State:   val.State,
		}
		items = append(items, item)

	}
	return items
}
func (l KernelNeighbor) List(w http.ResponseWriter, r *http.Request) {
	items := ListNeighbrs()
	ResponseJson(w, items)
}

type Device struct {
}

func (h Device) Router(router *mux.Router) {
	router.HandleFunc("/api/device", h.List).Methods("GET")
}

func ListDevices() []schema.Device {
	values := make([]schema.Device, 0, 1024)
	for u := range cache.Network.List() {
		if u == nil {
			break
		}
		c := u.Config.(*co.Network)
		if c == nil || c.Bridge == nil {
			continue
		}

		// Bridge device
		br := cn.Bridges.Get(c.Bridge.Name)
		if br != nil {
			name := br.L3Name()
			sts := cn.GetDevInfo(name)
			values = append(values, schema.Device{
				Network: c.Name,
				Name:    name,
				Address: cn.GetDevAddr(name),
				Mtu:     sts.Mtu,
				Mac:     sts.Mac,
				Recv:    sts.Recv,
				Send:    sts.Send,
				Drop:    sts.Drop,
				State:   sts.State,
			})
		}
		// OpenVPN device
		if c.OpenVPN != nil {
			name := c.OpenVPN.Device
			sts := cn.GetDevInfo(name)
			values = append(values, schema.Device{
				Network: c.Name,
				Name:    name,
				Address: cn.GetDevAddr(name),
				Mtu:     sts.Mtu,
				Mac:     sts.Mac,
				Recv:    sts.Recv,
				Send:    sts.Send,
				Drop:    sts.Drop,
				State:   sts.State,
			})
		}
	}
	// Output devices
	for l := range cache.Output.ListAll() {
		if l == nil {
			break
		}
		name := l.Device
		sts := cn.GetDevInfo(name)
		values = append(values, schema.Device{
			Network: l.Network,
			Name:    name,
			Mtu:     sts.Mtu,
			Mac:     sts.Mac,
			Recv:    sts.Recv,
			Send:    sts.Send,
			Drop:    sts.Drop,
			State:   sts.State,
		})
	}

	// Access devices
	for a := range cache.Access.List() {
		if a == nil {
			break
		}
		name := a.IfName
		sts := cn.GetDevInfo(name)
		values = append(values, schema.Device{
			Network: a.Network,
			Name:    name,
			Mtu:     sts.Mtu,
			Mac:     sts.Mac,
			Recv:    sts.Recv,
			Send:    sts.Send,
			Drop:    sts.Drop,
			State:   sts.State,
		})
	}

	// Physical links
	for _, d := range libol.ListPhyLinks() {
		values = append(values, schema.Device{
			Network: "-",
			Name:    d.Name,
			Mtu:     d.Mtu,
			Mac:     d.Mac,
			Recv:    d.Recv,
			Send:    d.Send,
			Drop:    d.Drop,
			State:   d.State,
			Address: cn.GetDevAddr(d.Name),
		})
	}
	for k, v := range values {
		device := &values[k]
		speed := schema.Speed{
			Name: v.Name,
			Send: v.Send,
			Recv: v.Recv,
		}
		device.RxSpeed, device.TxSpeed = cache.Speed.Out(speed)
	}

	return values
}
func (h Device) List(w http.ResponseWriter, r *http.Request) {
	dev := ListDevices()
	sort.SliceStable(dev, func(i, j int) bool {
		return dev[i].ID() > dev[j].ID()
	})
	ResponseJson(w, dev)
}

type Log struct {
}

func (l Log) Router(router *mux.Router) {
	router.HandleFunc("/api/log", l.List).Methods("GET")
	router.HandleFunc("/api/log", l.Add).Methods("POST")
}

func (l Log) List(w http.ResponseWriter, r *http.Request) {
	log := schema.NewLogSchema()
	ResponseJson(w, log)
}

func (l Log) Add(w http.ResponseWriter, r *http.Request) {
	log := &schema.Log{}
	if err := GetData(r, log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	libol.SetLevel(log.Level)

	ResponseMsg(w, 0, "")
}

type LDAP struct {
	cs SwitchApi
}

func (l LDAP) Router(router *mux.Router) {
	router.HandleFunc("/api/ldap", l.List).Methods("GET")
	router.HandleFunc("/api/ldap", l.Add).Methods("POST")
	router.HandleFunc("/api/ldap", l.Del).Methods("DELETE")
}

func (l LDAP) List(w http.ResponseWriter, r *http.Request) {
	config := l.cs.Config()
	if config != nil && config.Ldap != nil {
		cfg := config.Ldap
		value := schema.LDAP{
			BaseDN:    cfg.BaseDN,
			Attribute: cfg.Attribute,
			BindDN:    cfg.BindDN,
			BindPass:  cfg.BindPass,
			Server:    cfg.Server,
			EnableTls: cfg.Tls,
			Filter:    cfg.Filter,
			State:     cache.User.LDAPState(),
		}
		ResponseJson(w, value)
		return
	}
	ResponseJson(w, nil)
}

func (l LDAP) Add(w http.ResponseWriter, r *http.Request) {
	value := schema.LDAP{}
	if err := GetData(r, &value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l.cs.AddLDAP(value)
	ResponseMsg(w, 0, "")
}

func (l LDAP) Del(w http.ResponseWriter, r *http.Request) {
	l.cs.DelLDAP()
	ResponseMsg(w, 0, "")
}
