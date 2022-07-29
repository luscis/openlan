package api

import (
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	"net"
	"net/http"
	"time"
)

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
