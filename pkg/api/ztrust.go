package api

import (
	"net/http"
	"strings"

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
	router.HandleFunc("/api/network/{id}/guest/{user}", h.ListGuest).Methods("GET")
	router.HandleFunc("/api/network/{id}/guest/{user}", h.AddGuest).Methods("POST")
	router.HandleFunc("/api/network/{id}/guest/{user}", h.DelGuest).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/guest/{user}/knock", h.ListKnock).Methods("GET")
	router.HandleFunc("/api/network/{id}/guest/{user}/knock", h.AddKnock).Methods("POST")
}

func CheckUser(r *http.Request) (bool, string) {
	user, _, _ := r.BasicAuth()
	if strings.Contains(user, "@") {
		return false, strings.SplitN(user, "@", 2)[0]
	}
	return true, ""
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

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusBadRequest)
		return
	}

	admin, name := CheckUser(r)
	guests := make([]schema.ZGuest, 0, 1024)
	ztrust.ListGuest(func(obj schema.ZGuest) {
		if !admin {
			if obj.Name == name {
				guests = append(guests, obj)
			}
		} else {
			guests = append(guests, obj)
		}

	})

	ResponseJson(w, guests)
}

func (h ZTrust) AddGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusBadRequest)
		return
	}

	guest := &schema.ZGuest{}
	if err := GetData(r, guest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := vars["user"]
	admin, name := CheckUser(r)
	if !admin && user != name {
		http.Error(w, "have no permission", http.StatusForbidden)
		return
	}

	guest.Name = user
	if guest.Address == "" {
		client := cache.VPNClient.Get(id, guest.Name)
		if client != nil {
			guest.Address = client.Address
			guest.Device = client.Device
		}
	}
	if guest.Address == "" {
		http.Error(w, "can't find address", http.StatusBadRequest)
		return
	}

	libol.Debug("ZTrust.AddGuest %s@%s", guest.Name, id)
	if err := ztrust.AddGuest(guest.Name, guest.Address); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h ZTrust) DelGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusBadRequest)
		return
	}

	guest := &schema.ZGuest{}
	if err := GetData(r, guest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := vars["user"]
	admin, name := CheckUser(r)
	if !admin && user != name {
		http.Error(w, "have no permission", http.StatusForbidden)
		return
	}

	guest.Name = user
	libol.Debug("ZTrust.DelGuest %s@%s", guest.Name, id)
	if err := ztrust.DelGuest(guest.Name, guest.Address); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h ZTrust) ListKnock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusBadRequest)
		return
	}

	user := vars["user"]
	admin, name := CheckUser(r)
	if !admin && user != name {
		http.Error(w, "have no permission", http.StatusForbidden)
		return
	}

	rules := make([]schema.KnockRule, 0, 1024)
	ztrust.ListKnock(user, func(obj schema.KnockRule) {
		rules = append(rules, obj)
	})

	ResponseJson(w, rules)
}

func (h ZTrust) AddKnock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	ztrust := worker.ZTruster()
	if ztrust == nil {
		http.Error(w, "ZTrust disabled", http.StatusBadRequest)
		return
	}

	rule := &schema.KnockRule{}
	if err := GetData(r, rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := vars["user"]
	admin, name := CheckUser(r)
	if !admin && user != name {
		http.Error(w, "have no permission", http.StatusForbidden)
		return
	}

	libol.Debug("ZTrust.AddKnock %s@%s", user, id)
	if err := ztrust.Knock(user, rule.Protocol, rule.Dest, rule.Port, rule.Age); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
