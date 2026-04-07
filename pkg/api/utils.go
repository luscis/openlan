package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetData(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
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

func WriteAttachment(w http.ResponseWriter, file string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file))
}

func GetServer(r *http.Request) string {
	return strings.SplitN(r.Host, ":", 2)[0]
}
