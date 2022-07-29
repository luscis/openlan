package api

import (
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	"net/http"
)

type Neighbor struct {
}

func (h Neighbor) Router(router *mux.Router) {
	router.HandleFunc("/api/neighbor", h.List).Methods("GET")
}

func (h Neighbor) List(w http.ResponseWriter, r *http.Request) {
	neighbors := make([]schema.Neighbor, 0, 1024)
	for n := range cache.Neighbor.List() {
		if n == nil {
			break
		}
		neighbors = append(neighbors, models.NewNeighborSchema(n))
	}
	ResponseJson(w, neighbors)
}
