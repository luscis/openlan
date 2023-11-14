package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"gopkg.in/yaml.v2"
)

func ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.Marshal(v)
	if err == nil {
		libol.Debug("ResponseJson: %s", str)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ResponseMsg(w http.ResponseWriter, code int, message string) {
	ret := &schema.Message{
		Code:    code,
		Message: message,
	}
	ResponseJson(w, ret)
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

func GetData(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil
}

func GetQueryOne(req *http.Request, name string) string {
	query := req.URL.Query()
	if values, ok := query[name]; ok {
		return values[0]
	}
	return ""
}
