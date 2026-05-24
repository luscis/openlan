package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type RateLimit struct {
	cs SwitchApi
}

func (h RateLimit) Router(router *mux.Router) {
	router.HandleFunc("/api/interface/{id}/rate", h.Post).Methods("POST")
	router.HandleFunc("/api/interface/{id}/rate", h.Delete).Methods("DELETE")
}

func (h RateLimit) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	device := vars["id"]

	rate := &schema.Rate{}
	if err := GetData(r, rate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.cs.AddRate(device, rate.Speed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h RateLimit) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	device := vars["id"]
	if err := h.cs.DelRate(device); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
