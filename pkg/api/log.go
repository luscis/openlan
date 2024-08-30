package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type Log struct {
}

func (l Log) Router(router *mux.Router) {
	router.HandleFunc("/api/log", l.List).Methods("GET")
	router.HandleFunc("/api/log", l.Add).Methods("POST")
}

func (l Log) List(w http.ResponseWriter, r *http.Request) {
	log := schema.NewLogSchema()
	ResponseJson(w, log)
}

func (l Log) Add(w http.ResponseWriter, r *http.Request) {
	log := &schema.Log{}
	if err := GetData(r, log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	libol.SetLevel(log.Level)

	ResponseMsg(w, 0, "")
}
