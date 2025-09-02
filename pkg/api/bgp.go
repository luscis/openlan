package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type Bgp struct {
	Switcher Switcher
}

func (h Bgp) Router(router *mux.Router) {
	router.HandleFunc("/api/network/bgp", h.Get).Methods("GET")
	router.HandleFunc("/api/network/bgp", h.Post).Methods("POST")
	router.HandleFunc("/api/network/bgp", h.Remove).Methods("DELETE")
	router.HandleFunc("/api/network/bgp/neighbor", h.RemoveNeighbor).Methods("DELETE")
	router.HandleFunc("/api/network/bgp/neighbor", h.AddNeighbor).Methods("POST")
}

func (h Bgp) Get(w http.ResponseWriter, r *http.Request) {
	libol.Debug("Bgp.Get %s")
	if Call.bgper == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	data := Call.bgper.Get()
	ResponseJson(w, data)
}

func (h Bgp) Post(w http.ResponseWriter, r *http.Request) {
	data := schema.Bgp{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgper == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgper.Enable(data)
	ResponseMsg(w, 0, "")
}

func (h Bgp) Remove(w http.ResponseWriter, r *http.Request) {
	if Call.bgper == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgper.Disable()
	ResponseMsg(w, 0, "")
}

func (h Bgp) RemoveNeighbor(w http.ResponseWriter, r *http.Request) {
	nei := schema.BgpNeighbor{}
	if err := GetData(r, &nei); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgper == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgper.DelNeighbor(nei)
	ResponseMsg(w, 0, "")
}

func (h Bgp) AddNeighbor(w http.ResponseWriter, r *http.Request) {
	nei := schema.BgpNeighbor{}
	if err := GetData(r, &nei); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.bgper == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.bgper.AddNeighbor(nei)
	ResponseMsg(w, 0, "")
}
