package api

import (
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	"net/http"
)

type OnLine struct {
}

func (h OnLine) Router(router *mux.Router) {
	router.HandleFunc("/api/online", h.List).Methods("GET")
}

func (h OnLine) List(w http.ResponseWriter, r *http.Request) {
	nets := make([]schema.OnLine, 0, 1024)
	for u := range cache.Online.List() {
		if u == nil {
			break
		}
		nets = append(nets, models.NewOnLineSchema(u))
	}
	ResponseJson(w, nets)
}
