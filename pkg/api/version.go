package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/schema"
)

type Version struct {
}

func (l Version) Router(router *mux.Router) {
	router.HandleFunc("/api/version", l.List).Methods("GET")
}

func (l Version) List(w http.ResponseWriter, r *http.Request) {
	ver := schema.NewVersionSchema()
	ver.Expire = cache.User.ExpireTime()
	ResponseJson(w, ver)
}
