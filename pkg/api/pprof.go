package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"io/ioutil"
	"net/http"
)

var pprof = &libol.PProf{}

type PProf struct {
}

func (h PProf) Router(router *mux.Router) {
	router.HandleFunc("/api/pprof", h.Get).Methods("GET")
	router.HandleFunc("/api/pprof", h.Add).Methods("POST")
}

func (h PProf) Get(w http.ResponseWriter, r *http.Request) {
	pp := schema.PProf{Listen: pprof.Listen}
	ResponseJson(w, pp)
}

func (h PProf) Add(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pprof.Listen != "" && pprof.Error == nil {
		http.Error(w, "already listen on "+pprof.Listen, http.StatusConflict)
		return
	}

	pp := &schema.PProf{}
	if err := json.Unmarshal(body, pp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pprof.Listen = pp.Listen
	pprof.Start()
	ResponseMsg(w, 0, "")
}
