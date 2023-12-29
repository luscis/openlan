package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
)

type OpenAPI struct {
}

func (h OpenAPI) Router(router *mux.Router) {
	router.HandleFunc("/openvpn-api/profile", h.Get).Methods("HEAD")
	router.HandleFunc("/rest/{action}", h.Rest).Methods("GET")
}

func (h OpenAPI) Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("TODO"))
}

func GetNetwork(name string) string {
	values := strings.SplitN(name, "@", 2)
	if len(values) < 2 {
		return "default"
	}
	return values[1]
}

func (h OpenAPI) Error(w http.ResponseWriter, kind, message string) {
	w.Header().Set("Content-Type", "text/xml")
	context := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Error>
	<Type>%s</Type>
	<Synopsis>REST method failed</Synopsis>
	<Message>%s</Message>
</Error>
`, kind, message)
	_, _ = w.Write([]byte(context))
}

func (h OpenAPI) Rest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	user, pass, ok := r.BasicAuth()
	if !ok {
		h.Error(w, "Authorization Required", "AUTH: not have auth")
		return
	}
	if UserCheck(user, pass) != nil {
		h.Error(w, "Authorization Required", "AUTH: wrong username or password")
		return
	}

	name := GetNetwork(user)
	server := strings.SplitN(r.Host, ":", 2)[0]

	data, _ := cache.VPNClient.GetClientProfile(name, server)

	action := vars["action"]
	if action == "GetUserlogin" {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(data))
	} else {
		h.Error(w, "Internal Server Error", "ACTION: not support "+action)
		return
	}
}
