package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type FindHop struct {
	cs SwitchApi
}

func (rt FindHop) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/findhop", rt.List).Methods("GET")
	router.HandleFunc("/api/network/{id}/findhop", rt.Add).Methods("POST")
	router.HandleFunc("/api/network/{id}/findhop", rt.Del).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/findhop", rt.Save).Methods("PUT")
}

func (rt FindHop) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	routes := make([]schema.FindHop, 0, 1024)
	hoper := worker.FindHoper()

	hoper.ListHop(func(obj schema.FindHop) {
		routes = append(routes, obj)
	})
	ResponseJson(w, routes)
}

func (rt FindHop) Add(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	data := schema.FindHop{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hoper := worker.FindHoper()
	if err := hoper.AddHop(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ResponseJson(w, true)
}

func (rt FindHop) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	data := schema.FindHop{}
	if err := GetData(r, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hoper := worker.FindHoper()
	if err := hoper.DelHop(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ResponseJson(w, true)
}

func (rt FindHop) Save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	hoper := worker.FindHoper()
	hoper.SaveHop()

	ResponseJson(w, true)

}
