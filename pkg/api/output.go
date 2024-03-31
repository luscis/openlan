package api

import (
	"fmt"
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
	router.HandleFunc("/api/network/{id}/output", h.Delete).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/output", h.Save).Methods("PUT")
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
	output := &schema.Output{}
	if err := GetData(r, output); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cs := h.Switcher.Config()
	if cs.Network == nil {
		http.Error(w, "switch has no network can not add output", http.StatusBadRequest)
		return
	}
	network := cs.GetNetwork(output.Network)
	if network == nil {
		http.Error(w, fmt.Sprintf("switch has no network with %s can not add output", output.Network), http.StatusBadRequest)
		return
	}
	worker := GetWorker(output.Network)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	worker.AddOutput(output.Segment, output.Protocol, output.Remote)
	ResponseMsg(w, 0, "")
}

func (h Output) Delete(w http.ResponseWriter, r *http.Request) {
	output := &schema.Output{}
	if err := GetData(r, output); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cs := h.Switcher.Config()
	if cs.Network == nil {
		http.Error(w, "switch has no network can not del output", http.StatusBadRequest)
		return
	}
	network := cs.GetNetwork(output.Network)
	if network == nil {
		http.Error(w, fmt.Sprintf("switch has no network with %s can not del output", output.Network), http.StatusBadRequest)
		return
	}
	worker := GetWorker(output.Network)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	worker.DelOutput(output.Device)
	ResponseMsg(w, 0, "")
}

func (h Output) Save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	worker.SaveOutput()

	ResponseJson(w, "success")
}
