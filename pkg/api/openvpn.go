package api

import (
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/schema"
	"net/http"
)

type VPNClient struct {
}

func (h VPNClient) Router(router *mux.Router) {
	router.HandleFunc("/api/vpn/client", h.List).Methods("GET")
	router.HandleFunc("/api/vpn/client/{id}", h.List).Methods("GET")
}

func (h VPNClient) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	clients := make([]schema.VPNClient, 0, 1024)
	if name == "" {
		for n := range cache.Network.List() {
			if n == nil {
				break
			}
			for client := range cache.VPNClient.List(n.Name) {
				if client == nil {
					break
				}
				clients = append(clients, *client)
			}
		}
	} else {
		for client := range cache.VPNClient.List(name) {
			if client == nil {
				break
			}
			clients = append(clients, *client)
		}
	}
	ResponseJson(w, clients)
}
