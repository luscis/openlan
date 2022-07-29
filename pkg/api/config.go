package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Config struct {
	Switcher Switcher
}

func (c Config) Router(router *mux.Router) {
	router.HandleFunc("/api/config", c.List).Methods("GET")
	router.HandleFunc("/api/config/reload", c.Reload).Methods("PUT")
	router.HandleFunc("/api/config/save", c.Save).Methods("PUT")
}

func (c Config) List(w http.ResponseWriter, r *http.Request) {
	format := GetQueryOne(r, "format")
	if format == "yaml" {
		ResponseYaml(w, c.Switcher.Config())
	} else {
		ResponseJson(w, c.Switcher.Config())
	}
}

func (c Config) Reload(w http.ResponseWriter, r *http.Request) {
	c.Switcher.Reload()
	ResponseMsg(w, 0, "success")
}

func (c Config) Save(w http.ResponseWriter, r *http.Request) {
	c.Switcher.Save()
	ResponseMsg(w, 0, "success")
}
