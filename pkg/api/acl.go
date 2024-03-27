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
}

func (h ACL) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
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

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}
	acl := worker.ACLer()

	rule := &schema.ACLRule{}
	if err := GetData(r, rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := acl.AddRule(rule); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h ACL) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusInternalServerError)
		return
	}
	acl := worker.ACLer()

	rule := &schema.ACLRule{}
	if err := GetData(r, rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := acl.DelRule(rule); err == nil {
		ResponseJson(w, "success")
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
