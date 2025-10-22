package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type IPSec struct {
	cs SwitchApi
}

func (h IPSec) Router(router *mux.Router) {
	router.HandleFunc("/api/network/ipsec/tunnel", h.Get).Methods("GET")
	router.HandleFunc("/api/network/ipsec/tunnel", h.Post).Methods("POST")
	router.HandleFunc("/api/network/ipsec/tunnel", h.Delete).Methods("DELETE")
	router.HandleFunc("/api/network/ipsec/tunnel/restart", h.Start).Methods("PUT")
}

func (h IPSec) Get(w http.ResponseWriter, r *http.Request) {
	libol.Debug("IPSec.Get %s")
	tunnels := make([]schema.IPSecTunnel, 0, 1024)
	if Call.ipsecApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.ipsecApi.ListTunnels(func(obj schema.IPSecTunnel) {
		tunnels = append(tunnels, obj)
	})
	ResponseJson(w, tunnels)
}

func (h IPSec) Post(w http.ResponseWriter, r *http.Request) {
	tun := &schema.IPSecTunnel{}
	if err := GetData(r, tun); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.ipsecApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.ipsecApi.AddTunnel(*tun)
	ResponseMsg(w, 0, "")
}

func (h IPSec) Delete(w http.ResponseWriter, r *http.Request) {
	tun := &schema.IPSecTunnel{}
	if err := GetData(r, tun); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.ipsecApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.ipsecApi.DelTunnel(*tun)
	ResponseMsg(w, 0, "")
}

func (h IPSec) Start(w http.ResponseWriter, r *http.Request) {
	tun := &schema.IPSecTunnel{}
	if err := GetData(r, tun); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.ipsecApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.ipsecApi.StartTunnel(*tun)
	ResponseMsg(w, 0, "")
}
