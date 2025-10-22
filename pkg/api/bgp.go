package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type Bgp struct {
	cs SwitchApi
}

func (h Bgp) Router(router *mux.Router) {
	router.HandleFunc("/api/network/bgp", h.Get).Methods("GET")
	router.HandleFunc("/api/network/bgp", h.Post).Methods("POST")
	router.HandleFunc("/api/network/bgp", h.Remove).Methods("DELETE")
	router.HandleFunc("/api/network/bgp/neighbor", h.RemoveNeighbor).Methods("DELETE")
	router.HandleFunc("/api/network/bgp/neighbor", h.AddNeighbor).Methods("POST")
	router.HandleFunc("/api/network/bgp/advertis", h.RemoveAdvertis).Methods("DELETE")
	router.HandleFunc("/api/network/bgp/advertis", h.AddAdvertis).Methods("POST")
	router.HandleFunc("/api/network/bgp/receives", h.RemoveReceivess).Methods("DELETE")
	router.HandleFunc("/api/network/bgp/receives", h.AddReceivess).Methods("POST")
}

func (h Bgp) Get(w http.ResponseWriter, r *http.Request) {
	libol.Debug("Bgp.Get %s")
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	data := Call.bgpApi.Get()
	ResponseJson(w, data)
}

func (h Bgp) Post(w http.ResponseWriter, r *http.Request) {
	data := schema.Bgp{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.Enable(data)
	ResponseMsg(w, 0, "")
}

func (h Bgp) Remove(w http.ResponseWriter, r *http.Request) {
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.Disable()
	ResponseMsg(w, 0, "")
}

func (h Bgp) RemoveNeighbor(w http.ResponseWriter, r *http.Request) {
	nei := schema.BgpNeighbor{}
	if err := GetData(r, &nei); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.DelNeighbor(nei)
	ResponseMsg(w, 0, "")
}

func (h Bgp) AddNeighbor(w http.ResponseWriter, r *http.Request) {
	nei := schema.BgpNeighbor{}
	if err := GetData(r, &nei); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.AddNeighbor(nei)
	ResponseMsg(w, 0, "")
}

func (h Bgp) RemoveAdvertis(w http.ResponseWriter, r *http.Request) {
	data := schema.BgpPrefix{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.DelAdvertis(data)
	ResponseMsg(w, 0, "")
}

func (h Bgp) AddAdvertis(w http.ResponseWriter, r *http.Request) {
	data := schema.BgpPrefix{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.AddAdvertis(data)
	ResponseMsg(w, 0, "")
}

func (h Bgp) RemoveReceivess(w http.ResponseWriter, r *http.Request) {
	data := schema.BgpPrefix{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.DelReceives(data)
	ResponseMsg(w, 0, "")
}

func (h Bgp) AddReceivess(w http.ResponseWriter, r *http.Request) {
	data := schema.BgpPrefix{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgpApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgpApi.AddReceives(data)
	ResponseMsg(w, 0, "")
}
