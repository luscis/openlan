package libsock

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/luscis/openlan/pkg/libol"
	"golang.org/x/net/websocket"
)

type WsClient struct {
	Auth      libol.Auth
	Url       string
	TlsConfig *tls.Config
	Protocol  string
}

func (w *WsClient) Initialize() {
	u, _ := url.Parse(w.Url)
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
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
			"Authorization": {libol.BasicAuth(w.Auth.Username, w.Auth.Password)},
		}
	}
	return websocket.DialConfig(config)
}
