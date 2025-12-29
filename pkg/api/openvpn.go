package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/schema"
)

type VPNClient struct {
}

func (h VPNClient) Router(router *mux.Router) {
	router.HandleFunc("/api/vpn/client", h.List).Methods("GET")
	router.HandleFunc("/api/vpn/client/{id}", h.List).Methods("GET")
	router.HandleFunc("/api/vpn/client/{id}", h.Add).Methods("POST")
	router.HandleFunc("/api/vpn/client/{id}/kill", h.Kill).Methods("POST")
	router.HandleFunc("/api/vpn/client/{id}", h.Remove).Methods("DELETE")
}

func (h VPNClient) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	clients := make(map[string]schema.VPNClient, 1024)
	if name == "" {
		for n := range cache.Network.List() {
			if n == nil {
				break
			}
			for client := range cache.VPNClient.List(n.Name) {
				if client == nil {
					break
				}
				clients[client.Name] = *client
			}
		}
	} else {
		worker := Call.GetWorker(name)
		if worker == nil {
			http.Error(w, "Network not found", http.StatusBadRequest)
			return
		}

		for client := range cache.VPNClient.List(name) {
			if client == nil {
				break
			}
			clients[client.Name] = *client
		}

		worker.ListClients(func(name, address string) {
			if _, ok := clients[name]; ok {
				return
			}

			clients[name] = schema.VPNClient{
				Name:    name,
				Address: address,
			}
		})
	}

	items := make([]schema.VPNClient, 0, 1024)
	for _, v := range clients {
		items = append(items, v)
	}

	ResponseJson(w, items)
}

func (h VPNClient) Add(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	worker := Call.GetWorker(name)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	value := &schema.VPNClient{}
	if err := GetData(r, value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := worker.AddVPNClient(value.Name, value.Address); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (h VPNClient) Remove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	worker := Call.GetWorker(name)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	value := &schema.VPNClient{}
	if err := GetData(r, value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := worker.DelVPNClient(value.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (h VPNClient) Kill(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	worker := Call.GetWorker(name)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	value := &schema.VPNClient{}
	if err := GetData(r, value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := worker.KillVPNClient(value.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}
