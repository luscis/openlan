package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
)

type Output struct {
	Switcher Switcher
}

func (h Output) Router(router *mux.Router) {
	router.HandleFunc("/api/output", h.List).Methods("GET")
	router.HandleFunc("/api/output/{id}", h.Get).Methods("GET")
}

func (h Output) List(w http.ResponseWriter, r *http.Request) {
	outputs := make([]schema.Output, 0, 1024)
	for l := range cache.Output.List() {
		if l == nil {
			break
		}
		outputs = append(outputs, models.NewOutputSchema(l))
	}
	ResponseJson(w, outputs)
}

func (h Output) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	libol.Debug("Output.Get %s", vars["id"])
	output := cache.Output.Get(vars["id"])
	if output != nil {
		ResponseJson(w, models.NewOutputSchema(output))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}
