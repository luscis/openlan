package api

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

type KernelRoute struct {
}

func (l KernelRoute) Router(router *mux.Router) {
	router.HandleFunc("/api/kernel/route", l.List).Methods("GET")
}

func (l KernelRoute) List(w http.ResponseWriter, r *http.Request) {
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
		items = append(items, item)

	}
	ResponseJson(w, items)
}

type KernelNeighbor struct {
}

func (l KernelNeighbor) Router(router *mux.Router) {
	router.HandleFunc("/api/kernel/neighbor", l.List).Methods("GET")
}

func (l KernelNeighbor) List(w http.ResponseWriter, r *http.Request) {
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
	ResponseJson(w, items)
}

type Device struct {
}

func (h Device) Router(router *mux.Router) {
	router.HandleFunc("/api/device", h.List).Methods("GET")
	router.HandleFunc("/api/device/{id}", h.Get).Methods("GET")
}

func (h Device) List(w http.ResponseWriter, r *http.Request) {
	dev := make([]schema.Device, 0, 1024)
	for t := range network.Taps.List() {
		if t == nil {
			break
		}
		dev = append(dev, schema.Device{
			Name:     t.Name(),
			Mtu:      t.Mtu(),
			Provider: t.Type(),
		})
	}
	for t := range network.Bridges.List() {
		if t == nil {
			break
		}
		dev = append(dev, schema.Device{
			Name:     t.Name(),
			Mtu:      t.Mtu(),
			Provider: t.Type(),
		})
	}
	ResponseJson(w, dev)
}

func (h Device) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]
	if dev := network.Taps.Get(name); dev != nil {
		ResponseJson(w, schema.Device{
			Name:     dev.Name(),
			Mtu:      dev.Mtu(),
			Provider: dev.Type(),
		})
	} else if br := network.Bridges.Get(name); br != nil {
		now := time.Now().Unix()
		macs := make([]schema.HwMacInfo, 0, 32)
		for addr := range br.ListMac() {
			if addr == nil {
				break
			}
			macs = append(macs, schema.HwMacInfo{
				Address: net.HardwareAddr(addr.Address).String(),
				Device:  addr.Device.String(),
				Uptime:  now - addr.Uptime,
			})
		}
		slaves := make([]schema.Device, 0, 32)
		for dev := range br.ListSlave() {
			if dev == nil {
				break
			}
			slaves = append(slaves, schema.Device{
				Name:     dev.Name(),
				Mtu:      dev.Mtu(),
				Provider: dev.Type(),
			})
		}
		ResponseJson(w, schema.Bridge{
			Device: schema.Device{
				Name:     br.Name(),
				Mtu:      br.Mtu(),
				Provider: br.Type(),
			},
			Macs:   macs,
			Slaves: slaves,
			Stats:  br.Stats(),
		})
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
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

type Ldap struct {
	cs SwitchApi
}

func (l Ldap) Router(router *mux.Router) {
	router.HandleFunc("/api/ldap", l.List).Methods("GET")
	router.HandleFunc("/api/ldap", l.Add).Methods("POST")
	router.HandleFunc("/api/ldap", l.Del).Methods("DELETE")
}

func (l Ldap) List(w http.ResponseWriter, r *http.Request) {
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
			State:     cache.User.LdapState(),
		}
		ResponseJson(w, value)
		return
	}
	ResponseJson(w, nil)
}

func (l Ldap) Add(w http.ResponseWriter, r *http.Request) {
	value := schema.LDAP{}
	if err := GetData(r, &value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l.cs.AddLdap(value)
	ResponseMsg(w, 0, "")
}

func (l Ldap) Del(w http.ResponseWriter, r *http.Request) {
	l.cs.DelLdap()
	ResponseMsg(w, 0, "")
}
