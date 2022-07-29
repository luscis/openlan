package api

import (
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
	"net/http"
)

type Version struct {
}

func (l Version) Router(router *mux.Router) {
	router.HandleFunc("/api/version", l.List).Methods("GET")
}

func (l Version) List(w http.ResponseWriter, r *http.Request) {
	ver := schema.NewVersionSchema()
	ResponseJson(w, ver)
}
