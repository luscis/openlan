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
	h.cs.AddRate(device, rate.Speed)
}

func (h RateLimit) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	device := vars["id"]
	h.cs.DelRate(device)
}
