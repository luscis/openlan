package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
)

type Network struct {
	cs SwitchApi
}

func (h Network) Router(router *mux.Router) {
	router.HandleFunc("/api/network", h.List).Methods("GET")
	router.HandleFunc("/api/network", h.Post).Methods("POST")
	router.HandleFunc("/api/network", h.Save).Methods("PUT")
	router.HandleFunc("/api/network/{id}", h.Get).Methods("GET")
	router.HandleFunc("/api/network/{id}", h.Delete).Methods("DELETE")
	router.HandleFunc("/get/network/{id}/ovpn", h.Profile).Methods("GET")
	router.HandleFunc("/api/network/{id}/ovpn", h.Profile).Methods("GET")
	router.HandleFunc("/api/network/{id}/openvpn/restart", h.StartVPN).Methods("POST")
}

func (h Network) List(w http.ResponseWriter, r *http.Request) {
	nets := make([]schema.Network, 0, 1024)
	for u := range cache.Network.List() {
		if u == nil {
			break
		}
		nets = append(nets, models.NewNetworkSchema(u))
	}
	ResponseJson(w, nets)
}

func (h Network) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	net := cache.Network.Get(vars["id"])
	if net != nil {
		ResponseJson(w, models.NewNetworkSchema(net))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h Network) Post(w http.ResponseWriter, r *http.Request) {
	network := &schema.Network{}
	if err := GetData(r, network); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := libol.Marshal(&network.Config, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cs := h.cs.Config()
	obj, err := cs.AddNetwork(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cs.CorrectNetwork(obj, "json")
	if obj := cs.GetNetwork(obj.Name); obj != nil {
		h.cs.AddNetwork(obj.Name)
	} else {
		http.Error(w, obj.Name+" not found", http.StatusBadRequest)
		return
	}

	ResponseJson(w, "success")
}

func (h Network) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	network := vars["id"]
	worker := Call.GetWorker(network)
	if worker == nil {
		http.Error(w, "network not found", http.StatusBadRequest)
		return
	}
	h.cs.DelNetwork(network)
	ResponseJson(w, "success")
}

func (h Network) Save(w http.ResponseWriter, r *http.Request) {
	network := &schema.Network{}
	if err := GetData(r, network); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.cs.SaveNetwork(network.Name)
	ResponseJson(w, "success")
}

func (h Network) Profile(w http.ResponseWriter, r *http.Request) {
	server := strings.SplitN(r.Host, ":", 2)[0]
	vars := mux.Vars(r)
	data, err := cache.VPNClient.GetClientProfile(vars["id"], server)
	if err == nil {
		_, _ = w.Write([]byte(data))
	} else {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

func (h Network) StartVPN(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	worker := Call.GetWorker(id)
	if worker == nil {
		http.Error(w, "Network not found", http.StatusBadRequest)
		return
	}

	worker.StartVPN()
	ResponseJson(w, true)
}

type SNAT struct {
	cs SwitchApi
}

func (h SNAT) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/snat", h.Post).Methods("POST")
	router.HandleFunc("/api/network/{id}/snat", h.Delete).Methods("DELETE")
}

func (h SNAT) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	if obj := Call.GetWorker(name); obj != nil {
		obj.EnableSnat()
	} else {
		http.Error(w, name+" not found", http.StatusBadRequest)
		return
	}

	ResponseJson(w, "success")
}

func (h SNAT) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	if obj := Call.GetWorker(name); obj != nil {
		obj.DisableSnat()
	} else {
		http.Error(w, name+" not found", http.StatusBadRequest)
		return
	}

	ResponseJson(w, "success")
}

type DNAT struct {
	cs SwitchApi
}

func (h DNAT) Router(router *mux.Router) {
	router.HandleFunc("/api/network/{id}/dnat", h.Get).Methods("GET")
	router.HandleFunc("/api/network/{id}/dnat", h.Post).Methods("POST")
	router.HandleFunc("/api/network/{id}/dnat", h.Delete).Methods("DELETE")
}

func (h DNAT) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	caller := Call.GetWorker(name)
	if caller == nil {
		http.Error(w, name+" not found", http.StatusBadRequest)
		return
	}

	var items []schema.DNAT
	caller.ListDnat(func(data schema.DNAT) {
		items = append(items, data)
	})
	ResponseJson(w, items)
}

func (h DNAT) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	caller := Call.GetWorker(name)
	if caller == nil {
		http.Error(w, name+" not found", http.StatusBadRequest)
		return
	}

	value := schema.DNAT{}
	if err := GetData(r, &value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := caller.AddDnat(value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (h DNAT) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	caller := Call.GetWorker(name)
	if caller == nil {
		http.Error(w, name+" not found", http.StatusBadRequest)
		return
	}

	value := schema.DNAT{}
	if err := GetData(r, &value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := caller.DelDnat(value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

type RouterTunnel struct {
	cs SwitchApi
}

func (h RouterTunnel) Router(router *mux.Router) {
	router.HandleFunc("/api/network/router/tunnel", h.Post).Methods("POST")
	router.HandleFunc("/api/network/router/tunnel", h.Delete).Methods("DELETE")
}

func (h RouterTunnel) Post(w http.ResponseWriter, r *http.Request) {
	caller := Call.routerApi
	if caller == nil {
		http.Error(w, "Router not found", http.StatusBadRequest)
		return
	}

	value := schema.RouterTunnel{}
	if err := GetData(r, &value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := caller.AddTunnel(value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}

func (h RouterTunnel) Delete(w http.ResponseWriter, r *http.Request) {
	caller := Call.routerApi
	if caller == nil {
		http.Error(w, "Router not found", http.StatusBadRequest)
		return
	}

	value := schema.RouterTunnel{}
	if err := GetData(r, &value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := caller.DelTunnel(value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ResponseJson(w, "success")
}
