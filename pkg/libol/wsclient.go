package libol

import (
	"crypto/tls"
	"encoding/base64"
	"golang.org/x/net/websocket"
	"net/http"
	"net/url"
)

type Auth struct {
	Type     string
	Username string
	Password string
}

func BasicAuth(username, password string) string {
	auth := username + ":"
	if password != "" {
		auth += password
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

type WsClient struct {
	Auth      Auth
	Url       string
	TlsConfig *tls.Config
	Protocol  string
}

func (w *WsClient) Initialize() {
	u, _ := url.Parse(w.Url)
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	w.Url = u.String()
	w.TlsConfig = &tls.Config{InsecureSkipVerify: true}
}

func (w *WsClient) Dial() (ws *websocket.Conn, err error) {
	config, err := websocket.NewConfig(w.Url, w.Url)
	if err != nil {
		return nil, err
	}
	if w.Protocol != "" {
		config.Protocol = []string{w.Protocol}
	}
	config.TlsConfig = w.TlsConfig
	if w.Auth.Type == "basic" {
		config.Header = http.Header{
			"Authorization": {BasicAuth(w.Auth.Username, w.Auth.Password)},
		}
	}
	return websocket.DialConfig(config)
}
