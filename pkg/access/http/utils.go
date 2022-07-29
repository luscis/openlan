package http

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"net/http"
)

func ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ResponseYaml(w http.ResponseWriter, v interface{}) {
	str, err := yaml.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetQueryOne(req *http.Request, name string) string {
	query := req.URL.Query()
	if values, ok := query[name]; ok {
		return values[0]
	}
	return ""
}
