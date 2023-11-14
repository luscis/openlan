package api

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
)

type User struct {
}

func (h User) Router(router *mux.Router) {
	router.HandleFunc("/api/user", h.List).Methods("GET")
	router.HandleFunc("/api/user", h.Add).Methods("POST")
	router.HandleFunc("/api/user/{id}", h.Get).Methods("GET")
	router.HandleFunc("/api/user/{id}", h.Add).Methods("POST")
	router.HandleFunc("/api/user/{id}", h.Del).Methods("DELETE")
	router.HandleFunc("/api/user/{id}/check", h.Check).Methods("POST")
}

func (h User) List(w http.ResponseWriter, r *http.Request) {
	users := make([]schema.User, 0, 1024)
	for u := range cache.User.List() {
		if u == nil {
			break
		}
		users = append(users, models.NewUserSchema(u))
	}
	sort.SliceStable(users, func(i, j int) bool {
		return users[i].Network+users[i].Name > users[j].Network+users[j].Name
	})
	ResponseJson(w, users)
}

func (h User) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := cache.User.Get(vars["id"])
	if user != nil {
		ResponseJson(w, models.NewUserSchema(user))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h User) Add(w http.ResponseWriter, r *http.Request) {
	user := &schema.User{}
	if err := GetData(r, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cache.User.Add(models.SchemaToUserModel(user))
	if err := cache.User.Save(); err != nil {
		libol.Warn("AddUser %s", err)
	}
	ResponseMsg(w, 0, "")
}

func (h User) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("DelUser %s", vars["id"])

	cache.User.Del(vars["id"])
	if err := cache.User.Save(); err != nil {
		libol.Warn("DelUser %s", err)
	}
	ResponseMsg(w, 0, "")
}

func (h User) Check(w http.ResponseWriter, r *http.Request) {
	user := &schema.User{}
	if err := GetData(r, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.SchemaToUserModel(user)
	if _, err := cache.User.Check(model); err == nil {
		ResponseMsg(w, 0, "success")
	} else {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
}
