package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type Ceci struct {
	cs SwitchApi
}

func (h Ceci) Router(router *mux.Router) {
	router.HandleFunc("/api/network/ceci/tcp", h.Get).Methods("GET")
	router.HandleFunc("/api/network/ceci/tcp", h.Post).Methods("POST")
	router.HandleFunc("/api/network/ceci/tcp", h.Remove).Methods("DELETE")
}

func (h Ceci) Get(w http.ResponseWriter, r *http.Request) {
	libol.Debug("Ceci.Get %s")
	ResponseJson(w, nil)
}

func (h Ceci) Post(w http.ResponseWriter, r *http.Request) {
	data := schema.CeciTcp{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.ceciApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.ceciApi.AddTcp(data)
	ResponseMsg(w, 0, "")
}

func (h Ceci) Remove(w http.ResponseWriter, r *http.Request) {
	data := schema.CeciTcp{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if Call.ceciApi == nil {
		http.Error(w, "network is nil", http.StatusBadRequest)
		return
	}
	Call.ceciApi.DelTcp(data)
	ResponseMsg(w, 0, "")
}
