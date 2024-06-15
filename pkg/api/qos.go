package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type QosApi struct {
}

func (h QosApi) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/qos", h.List).Methods("GET")
	router.HandleFunc("/api/network/{id}/qos", h.Add).Methods("POST")
	router.HandleFunc("/api/network/{id}/qos", h.Del).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/qos", h.Save).Methods("PUT")
}

func (h QosApi) List(w http.ResponseWriter, r *http.Request) {

	qosList := make([]schema.Qos, 0, 1024)
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}

	var qos = worker.Qoser()
	qos.ListQosUsers(func(obj schema.Qos) {
		qosList = append(qosList, obj)
	})

	ResponseJson(w, qosList)
}

func (h QosApi) Add(w http.ResponseWriter, r *http.Request) {

	qos := &schema.Qos{}
	if err := GetData(r, qos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}

	if qos != nil {
		if err := worker.Qoser().AddQosUser(qos.Name, qos.InSpeed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ResponseJson(w, true)
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h QosApi) Del(w http.ResponseWriter, r *http.Request) {

	qos := &schema.Qos{}
	if err := GetData(r, qos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}

	if qos != nil {
		if err := worker.Qoser().DelQosUser(qos.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ResponseJson(w, true)
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h QosApi) Save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}
	qos := worker.Qoser()
	qos.Save()

	ResponseJson(w, "success")
}
