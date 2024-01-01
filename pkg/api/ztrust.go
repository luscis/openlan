package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type ZTrust struct {
	Switcher Switcher
}

func (h ZTrust) Router(router *mux.Router) {
	router.HandleFunc("/api/ztrust", h.List).Methods("GET")
	router.HandleFunc("/api/ztrust/{id}", h.Get).Methods("GET")
	router.HandleFunc("/api/ztrust/{id}/guest/{user}", h.GetGuest).Methods("GET")
	router.HandleFunc("/api/ztrust/{id}/guest/{user}", h.AddGuest).Methods("POST")
	router.HandleFunc("/api/ztrust/{id}/guest/{user}", h.DelGuest).Methods("DELETE")
	router.HandleFunc("/api/ztrust/{id}/guest/{user}/knock", h.ListKnock).Methods("GET")
	router.HandleFunc("/api/ztrust/{id}/guest/{user}/knock", h.AddKnock).Methods("POST")
}

func (h ZTrust) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, "TODO")
}

func (h ZTrust) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("ZTrust.GET %s", vars["id"])
	ResponseJson(w, "TODO")
}

func (h ZTrust) GetGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("ZTrust.AddGuest %s", vars["id"])
	ResponseJson(w, "TODO")
}

func (h ZTrust) AddGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusInternalServerError)
		return
	}

	guest := &schema.ZGuest{}
	if err := GetData(r, guest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	guest.Name = vars["user"]
	libol.Info("ZTrust.AddGuest %s@%s", guest.Name, id)

	if err := ztrust.AddGuest(guest.Name, guest.Address); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h ZTrust) DelGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusInternalServerError)
		return
	}

	guest := &schema.ZGuest{}
	if err := GetData(r, guest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	guest.Name = vars["user"]
	libol.Info("ZTrust.DelGuest %s@%s", guest.Name, id)
	if err := ztrust.DelGuest(guest.Name, guest.Address); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h ZTrust) ListKnock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("ZTrust.ListKnock %s", vars["id"])
	ResponseJson(w, "TODO")
}

func (h ZTrust) AddKnock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusInternalServerError)
		return
	}

	rule := &schema.KnockRule{}
	if err := GetData(r, rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := vars["user"]
	libol.Info("ZTrust.AddKnock %s@%s", rule.Name, id)

	if err := ztrust.Knock(name, rule.Protocl, rule.Dest, rule.Port, 0); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
