package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type Rate struct {
	cs SwitchApi
}

func (h Rate) Router(router *mux.Router) {
	router.HandleFunc("/api/interface/{id}/rate", h.Post).Methods("POST")
	router.HandleFunc("/api/interface/{id}/rate", h.Delete).Methods("DELETE")
}

func (h Rate) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	device := vars["id"]

	rate := &schema.Rate{}
	if err := GetData(r, rate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.cs.AddRate(device, rate.Speed)
}

func (h Rate) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	device := vars["id"]
	h.cs.DelRate(device)
}
