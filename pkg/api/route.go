package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type Route struct {
	cs SwitchApi
}

func (rt Route) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/route", rt.List).Methods("GET")
	router.HandleFunc("/api/network/{id}/route", rt.Add).Methods("POST")
	router.HandleFunc("/api/network/{id}/route", rt.Del).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/route", rt.Save).Methods("PUT")
}

func (rt Route) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	routes := make([]schema.PrefixRoute, 0, 1024)

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	worker.ListRoute(func(obj schema.PrefixRoute) {
		routes = append(routes, obj)
	})
	ResponseJson(w, routes)
}

func (rt Route) Add(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	pr := &schema.PrefixRoute{}
	if err := GetData(r, pr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := worker.AddRoute(pr, rt.cs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ResponseJson(w, true)
}

func (rt Route) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	pr := &schema.PrefixRoute{}
	if err := GetData(r, pr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := worker.DelRoute(pr, rt.cs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ResponseJson(w, true)
}

func (rt Route) Save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	worker.SaveRoute()

	ResponseJson(w, true)

}
