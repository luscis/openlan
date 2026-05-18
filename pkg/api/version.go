package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/schema"
)

type Version struct {
	cs SwitchApi
}

func (l Version) Router(router *mux.Router) {
	router.HandleFunc("/api/version", l.List).Methods("GET")
	router.HandleFunc("/api/version/cert", l.CertList).Methods("GET")
	router.HandleFunc("/api/version/cert", l.CertUpdate).Methods("POST")
	router.HandleFunc("/api/version/crypt", l.CryptList).Methods("GET")
	router.HandleFunc("/api/version/crypt", l.CryptUpdate).Methods("POST")
}

func (l Version) List(w http.ResponseWriter, r *http.Request) {
	ver := schema.NewVersionSchema()
	ce := l.cs.GetCert()
	ver.Expire = ce.CertExpire
	ResponseJson(w, ver)
}

func (l Version) CertList(w http.ResponseWriter, r *http.Request) {
	ce := l.cs.GetCert()
	ResponseJson(w, ce)
}

func (l Version) CertUpdate(w http.ResponseWriter, r *http.Request) {
	ce := schema.VersionCert{}
	if err := GetData(r, &ce); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.cs.UpdateCert(ce)
	ResponseJson(w, "success")
}

func (l Version) CryptList(w http.ResponseWriter, r *http.Request) {
	cp := l.cs.GetCrypt()
	ResponseJson(w, cp)
}

func (l Version) CryptUpdate(w http.ResponseWriter, r *http.Request) {
	cp := schema.SwitchCrypt{}
	if err := GetData(r, &cp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.cs.UpdateCrypt(cp)
	ResponseJson(w, "success")
}
