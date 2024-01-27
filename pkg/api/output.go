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
	router.HandleFunc("/api/network/{id}/output", h.Get).Methods("GET")
	router.HandleFunc("/api/network/{id}/output", h.Post).Methods("POST")
}

func (h Output) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	libol.Debug("Output.Get %s")
	outputs := make([]schema.Output, 0, 1024)
	for l := range cache.Output.List(name) {
		if l == nil {
			break
		}
		outputs = append(outputs, models.NewOutputSchema(l))
	}
	ResponseJson(w, outputs)
}

func (h Output) Post(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, "outputs")
}
