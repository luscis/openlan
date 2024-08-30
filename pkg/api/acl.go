package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type ACL struct {
	Switcher Switcher
}

func (h ACL) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/acl", h.List).Methods("GET")
	router.HandleFunc("/api/network/{id}/acl", h.Add).Methods("POST")
	router.HandleFunc("/api/network/{id}/acl", h.Del).Methods("DELETE")
	router.HandleFunc("/api/network/{id}/acl", h.Save).Methods("PUT")
}

func (h ACL) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	acl := worker.ACLer()

	rules := make([]schema.ACLRule, 0, 1024)
	acl.ListRules(func(obj schema.ACLRule) {
		rules = append(rules, obj)
	})

	ResponseJson(w, rules)
}

func (h ACL) Add(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	acl := worker.ACLer()

	rule := &schema.ACLRule{}
	if err := GetData(r, rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := acl.AddRule(rule); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h ACL) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	acl := worker.ACLer()

	rule := &schema.ACLRule{}
	if err := GetData(r, rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := acl.DelRule(rule); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h ACL) Save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}
	acl := worker.ACLer()
	acl.SaveRule()

	ResponseJson(w, "success")
}
