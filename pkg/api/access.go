package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
)

type Access struct {
}

func (h Access) Router(router *mux.Router) {
	router.HandleFunc("/api/point", h.List).Methods("GET")
	router.HandleFunc("/api/point/{network}", h.Get).Methods("GET")
}

func (h Access) List(w http.ResponseWriter, r *http.Request) {
	points := make([]schema.Access, 0, 1024)
	for u := range cache.Access.List() {
		if u == nil {
			break
		}
		points = append(points, models.NewAccessSchema(u))
	}
	ResponseJson(w, points)
}

func (h Access) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["network"]

	points := make([]schema.Access, 0, 1024)
	for u := range cache.Access.List() {
		if u == nil {
			break
		}
		if u.Network == name {
			points = append(points, models.NewAccessSchema(u))
		}
	}
	ResponseJson(w, points)
}
