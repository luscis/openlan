package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type ZTrust struct {
	Switcher Switcher
}

func (h ZTrust) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/ztrust", h.List).Methods("GET")
	router.HandleFunc("/api/network/{id}/guest", h.ListGuest).Methods("GET")
	router.HandleFunc("/api/network/{id}/guest/{user}", h.AddGuest).Methods("POST")
	router.HandleFunc("/api/network/{id}/guest/{user}", h.DelGuest).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/guest/{user}/knock", h.ListKnock).Methods("GET")
	router.HandleFunc("/api/network/{id}/guest/{user}/knock", h.AddKnock).Methods("POST")
}

func (h ZTrust) List(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, "TODO")
}

func (h ZTrust) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("ZTrust.GET %s", vars["id"])
	ResponseJson(w, "TODO")
}

func (h ZTrust) ListGuest(w http.ResponseWriter, r *http.Request) {
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

	guests := make([]schema.ZGuest, 0, 1024)
	ztrust.ListGuest(func(obj schema.ZGuest) {
		guests = append(guests, obj)
	})

	ResponseJson(w, guests)
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
	if guest.Address == "" {
		client := cache.VPNClient.Get(id, guest.Name)
		if client != nil {
			guest.Address = client.Address
			guest.Device = client.Device
		}
	}
	if guest.Address == "" {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

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

	name := vars["user"]
	rules := make([]schema.KnockRule, 0, 1024)
	ztrust.ListKnock(name, func(obj schema.KnockRule) {
		rules = append(rules, obj)
	})

	ResponseJson(w, rules)
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

	if err := ztrust.Knock(name, rule.Protocol, rule.Dest, rule.Port, rule.Age); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
