package v5

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/luscis/openlan/cmd/api"
	"github.com/luscis/openlan/pkg/libol"
)

type Client struct {
	Auth libol.Auth
	Host string
}

func (cl Client) NewRequest(url string) *libol.HttpClient {
	client := &libol.HttpClient{
		Auth: libol.Auth{
			Type:     "basic",
			Username: cl.Auth.Username,
			Password: cl.Auth.Password,
		},
		Url: url,
	}
	return client
}

func (cl Client) GetBody(url string) ([]byte, error) {
	client := cl.NewRequest(url)
	r, err := client.Do()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, libol.NewErr(r.Status)
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cl Client) JSON(client *libol.HttpClient, i, o interface{}) error {
	out := cl.Log()
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	out.Debug("Client.JSON -> %s %s", client.Method, client.Url)
	out.Debug("Client.JSON -> %s", string(data))
	client.Payload = bytes.NewReader(data)
	if r, err := client.Do(); err != nil {
		return err
	} else {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		out.Debug("client.JSON <- %s", string(body))
		if r.StatusCode != http.StatusOK {
			return libol.NewErr("%s %s", r.Status, body)
		} else if o != nil {
			if err := json.Unmarshal(body, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cl Client) GetJSON(url string, v interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "GET"
	return cl.JSON(client, nil, v)
}

func (cl Client) PostJSON(url string, i, o interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "POST"
	return cl.JSON(client, i, o)
}

func (cl Client) PutJSON(url string, i, o interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "PUT"
	return cl.JSON(client, i, o)
}

func (cl Client) DeleteJSON(url string, i, o interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "DELETE"
	return cl.JSON(client, i, o)
}

func (cl Client) Log() *libol.SubLogger {
	return libol.NewSubLogger("cli")
}

type Cmd struct {
}

func (c Cmd) NewHttp(token string) Client {
	values := strings.SplitN(token, ":", 2)
	username := values[0]
	password := values[0]
	if len(values) == 2 {
		password = values[1]
	}
	client := Client{
		Auth: libol.Auth{
			Username: username,
			Password: password,
		},
	}
	return client
}

func (c Cmd) Url(prefix, name string) string {
	return ""
}

func (c Cmd) Tmpl() string {
	return ""
}

func (c Cmd) Out(data interface{}, format string, tmpl string) error {
	if tmpl == "" {
		format = "yaml"
	}
	return api.Out(data, format, tmpl)
}

func (c Cmd) Log() *libol.SubLogger {
	return libol.NewSubLogger("cli")
}
