package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct {
	cs SwitchApi
}

func (c Config) Router(router *mux.Router) {
	router.HandleFunc("/api/config", c.List).Methods("GET")
	router.HandleFunc("/api/config/save", c.Save).Methods("PUT")
}

func (c Config) List(w http.ResponseWriter, r *http.Request) {
	format := GetQueryOne(r, "format")
	if format == "yaml" {
		ResponseYaml(w, c.cs.Config())
	} else {
		ResponseJson(w, c.cs.Config())
	}
}

func (c Config) Save(w http.ResponseWriter, r *http.Request) {
	c.cs.Save()
	ResponseMsg(w, 0, "success")
}
