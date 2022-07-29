package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type VxLAN struct {
	Switcher Switcher
}

func (l VxLAN) Router(router *mux.Router) {
	router.HandleFunc("/api/vxlan", l.List).Methods("GET")
	router.HandleFunc("/api/vxlan/{id}", l.List).Methods("GET")
}

func (l VxLAN) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, nil)
}
