package libol

import (
	"crypto/tls"
	"io"
	"net/http"
)

type HttpClient struct {
	Method    string
	Url       string
	Payload   io.Reader
	Auth      Auth
	TlsConfig *tls.Config
	Client    *http.Client
}

func (cl *HttpClient) Do() (*http.Response, error) {
	if cl.Method == "" {
		cl.Method = "GET"
	}
	if cl.TlsConfig == nil {
		cl.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := http.NewRequest(cl.Method, cl.Url, cl.Payload)
	if err != nil {
		return nil, err
	}
	if cl.Auth.Type == "basic" {
		req.Header.Set("Authorization", BasicAuth(cl.Auth.Username, cl.Auth.Password))
	}
	cl.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: cl.TlsConfig,
		},
	}
	return cl.Client.Do(req)
}

func (cl *HttpClient) Close() {
	if cl.Client != nil {
		cl.Client.CloseIdleConnections()
	}
}
