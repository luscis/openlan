package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/libol"
	"net/http"
)

type Http struct {
	pointer Pointer
	listen  string
	server  *http.Server
	crtFile string
	keyFile string
	pubDir  string
	router  *mux.Router
	token   string
}

func NewHttp(pointer Pointer) (h *Http) {
	h = &Http{
		pointer: pointer,
	}
	if config := pointer.Config(); config != nil {
		if config.Http != nil {
			h.listen = config.Http.Listen
			h.pubDir = config.Http.Public
		}
	}
	return h
}

func (h *Http) Initialize() {
	r := h.Router()
	if h.server == nil {
		h.server = &http.Server{
			Addr:    h.listen,
			Handler: r,
		}
	}
	h.token = libol.GenRandom(32)
	libol.Info("Http.Initialize: AdminToken: %s", h.token)
	h.LoadRouter()
}

func (h *Http) IsAuth(w http.ResponseWriter, r *http.Request) bool {
	token, pass, ok := r.BasicAuth()
	libol.Debug("Http.IsAuth token: %s, pass: %s", token, pass)
	if !ok || token != h.token {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Authorization Required.", http.StatusUnauthorized)
		return false
	}
	return true
}

func (h *Http) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.IsAuth(w, r) {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "Authorization Required.", http.StatusUnauthorized)
		}
	})
}

func (h *Http) Router() *mux.Router {
	if h.router == nil {
		h.router = mux.NewRouter()
		h.router.Use(h.Middleware)
	}
	return h.router
}

func (h *Http) LoadRouter() {
	router := h.Router()

	router.HandleFunc("/current/uuid", func(w http.ResponseWriter, r *http.Request) {
		format := GetQueryOne(r, "format")
		if format == "yaml" {
			ResponseYaml(w, h.pointer.UUID())
		} else {
			ResponseJson(w, h.pointer.UUID())
		}
	})
	router.HandleFunc("/current/config", func(w http.ResponseWriter, r *http.Request) {
		format := GetQueryOne(r, "format")
		if format == "yaml" {
			ResponseYaml(w, h.pointer.Config())
		} else {
			ResponseJson(w, h.pointer.Config())
		}
	})
}

func (h *Http) Start() {
	h.Initialize()
	libol.Info("Http.Start %s", h.listen)
	if h.keyFile == "" || h.crtFile == "" {
		if err := h.server.ListenAndServe(); err != nil {
			libol.Error("Http.Start on %s: %s", h.listen, err)
			return
		}
	}
}

func (h *Http) Shutdown() {
	libol.Info("Http.Shutdown %s", h.listen)
	if err := h.server.Shutdown(context.Background()); err != nil {
		// Error from closing listeners, or context timeout:
		libol.Error("Http.Shutdown: %v", err)
	}
}
