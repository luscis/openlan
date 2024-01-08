package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
)

type Link struct {
	Switcher Switcher
}

func (h Link) Router(router *mux.Router) {
	router.HandleFunc("/api/link", h.List).Methods("GET")
	router.HandleFunc("/api/link/{id}", h.Get).Methods("GET")
}

func (h Link) List(w http.ResponseWriter, r *http.Request) {
	links := make([]schema.Link, 0, 1024)
	for l := range cache.Link.List() {
		if l == nil {
			break
		}
		links = append(links, models.NewLinkSchema(l))
	}
	ResponseJson(w, links)
}

func (h Link) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	libol.Debug("Link.Get %s", vars["id"])
	link := cache.Link.Get(vars["id"])
	if link != nil {
		ResponseJson(w, models.NewLinkSchema(link))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}
