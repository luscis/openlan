package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type IPSec struct {
	Switcher Switcher
}

func (h IPSec) Router(router *mux.Router) {
	router.HandleFunc("/api/network/ipsec/tunnel", h.Get).Methods("GET")
	router.HandleFunc("/api/network/ipsec/tunnel", h.Post).Methods("POST")
	router.HandleFunc("/api/network/ipsec/tunnel", h.Delete).Methods("DELETE")
}

func (h IPSec) Get(w http.ResponseWriter, r *http.Request) {
	libol.Debug("IPSec.Get %s")
	tunnels := make([]schema.IPSecTunnel, 0, 1024)
	if Call.secer == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.secer.ListTunnels(func(obj schema.IPSecTunnel) {
		tunnels = append(tunnels, obj)
	})
	ResponseJson(w, tunnels)
}

func (h IPSec) Post(w http.ResponseWriter, r *http.Request) {
	tun := &schema.IPSecTunnel{}
	if err := GetData(r, tun); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if Call.secer == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.secer.AddTunnel(*tun)
	ResponseMsg(w, 0, "")
}

func (h IPSec) Delete(w http.ResponseWriter, r *http.Request) {
	tun := &schema.IPSecTunnel{}
	if err := GetData(r, tun); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if Call.secer == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.secer.DelTunnel(*tun)
	ResponseMsg(w, 0, "")
}
