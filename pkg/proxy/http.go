package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"gopkg.in/yaml.v2"
)

type HttpRecord struct {
	Count    int
	LastAt   string
	CreateAt string
	Domain   string
}

func (r *HttpRecord) Update() {
	if r.Count == 0 {
		r.CreateAt = time.Now().Local().String()
	}
	r.Count += 1
	r.LastAt = time.Now().Local().String()
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Oops!", http.StatusNotFound)
}

type HttpProxy struct {
	proxer   Proxyer
	pass     map[string]string
	out      *libol.SubLogger
	server   *http.Server
	cfg      *co.HttpProxy
	api      *mux.Router
	startat  time.Time
	requests map[string]*HttpRecord
	lock     sync.RWMutex
}

var (
	connectOkay = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
)

func decodeBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func encodeBasicAuth(value string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(value))
}

func encodeJson(w http.ResponseWriter, v interface{}) {
	str, err := json.Marshal(v)
	if err == nil {
		libol.Debug("ResponseJson: %s", str)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func encodeYaml(w http.ResponseWriter, v interface{}) {
	str, err := yaml.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func encodeText(w http.ResponseWriter, tmpl string, v interface{}) {
	obj := template.New("main")
	if tmpl, err := obj.Parse(tmpl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		if err := tmpl.Execute(w, v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func NewHttpProxy(cfg *co.HttpProxy, px Proxyer) *HttpProxy {
	h := &HttpProxy{
		out:      libol.NewSubLogger(cfg.Listen),
		cfg:      cfg,
		pass:     make(map[string]string),
		api:      mux.NewRouter(),
		requests: make(map[string]*HttpRecord),
		proxer:   px,
	}

	h.server = &http.Server{
		Addr:    cfg.Listen,
		Handler: h,
	}
	auth := cfg.Auth
	if auth != nil && auth.Username != "" {
		h.pass[auth.Username] = auth.Password
	}

	h.loadUrl()
	h.loadPass()

	return h
}

func (t *HttpProxy) loadUrl() {
	if strings.HasPrefix(t.cfg.Listen, "127.0.0.") {
		t.api.HandleFunc("/", t.GetStats).Methods("GET")
		t.api.HandleFunc("/api", t.GetApi).Methods("GET")
		t.api.HandleFunc("/api/config", t.GetConfig).Methods("GET")
		t.api.HandleFunc("/api/match/{domain}/to/{backend}", t.AddMatch).Methods("POST")
		t.api.HandleFunc("/api/match/{domain}/to/{backend}", t.DelMatch).Methods("DELETE")
		t.api.HandleFunc("/pac", t.GetPac).Methods("GET")
	}
	t.api.NotFoundHandler = http.HandlerFunc(NotFound)
}

func (t *HttpProxy) loadPass() {
	file := t.cfg.Password
	if file == "" {
		return
	}
	reader, err := libol.OpenRead(file)
	if err != nil {
		libol.Warn("HttpProxy.LoadPass open %v", err)
		return
	}
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.SplitN(line, ":", 2)
		if len(columns) < 2 {
			continue
		}
		user := columns[0]
		pass := columns[1]
		t.pass[user] = pass
	}
	if err := scanner.Err(); err != nil {
		libol.Warn("HttpProxy.LoadPass scaner %v", err)
	}
}

func (t *HttpProxy) isAuth(username, password string) bool {
	if p, ok := t.pass[username]; ok {
		return p == password
	}
	return false
}

func (t *HttpProxy) CheckAuth(w http.ResponseWriter, r *http.Request) bool {
	if len(t.pass) == 0 {
		return true
	}
	auth := r.Header.Get("Proxy-Authorization")
	user, password, ok := decodeBasicAuth(auth)
	if !ok || !t.isAuth(user, password) {
		w.Header().Set("Proxy-Authenticate", "Basic")
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return false
	}
	return true
}

func (t *HttpProxy) toDirect(w http.ResponseWriter, p *http.Response) {
	defer p.Body.Close()
	for key, value := range p.Header {
		if key == "Proxy-Authorization" {
			if len(value) > 0 { // Remove first value for next proxy.
				value = value[1:]
			}
		}
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(p.StatusCode)
	_, _ = io.Copy(w, p.Body)
}

func (t *HttpProxy) toTunnel(w http.ResponseWriter, conn net.Conn) {
	src, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer src.Close()
	wait := libol.NewWaitOne(2)
	libol.Go(func() {
		defer wait.Done()
		// The returned bufio.Reader may contain unprocessed buffered data from the client.
		// Copy them to dst so we can use src directly.
		if n := bio.Reader.Buffered(); n > 0 {
			n64, err := io.CopyN(conn, bio, int64(n))
			if n64 != int64(n) || err != nil {
				t.out.Warn("HttpProxy.tunnel io.CopyN:", n64, err)
				return
			}
		}
		if _, err := io.Copy(conn, src); err != nil {
			t.out.Debug("HttpProxy.tunnel from ws %s", err)
		}
	})
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(src, conn); err != nil {
			t.out.Debug("HttpProxy.tunnel from target %s", err)
		}
	})
	wait.Wait()
	t.out.Debug("HttpProxy.tunnel %s exit", conn.RemoteAddr())
}

func (t *HttpProxy) openConn(protocol, remote string, insecure bool) (net.Conn, error) {
	if protocol == "https" {
		conf := &tls.Config{
			InsecureSkipVerify: insecure,
		}
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		return tls.DialWithDialer(dialer, "tcp", remote, conf)

	}
	return net.DialTimeout("tcp", remote, 10*time.Second)
}

func (t *HttpProxy) cloneRequest(r *http.Request, secret string) ([]byte, error) {
	var err error
	var b bytes.Buffer

	reqURI := r.RequestURI
	if reqURI == "" {
		reqURI = r.URL.RequestURI()
	}

	fmt.Fprintf(&b, "%s %s HTTP/%d.%d\r\n", r.Method, reqURI, r.ProtoMajor, r.ProtoMinor)

	host := r.Host
	fmt.Fprintf(&b, "Host: %s\r\n", host)
	if secret != "" {
		fmt.Fprintf(&b, "Proxy-Authorization: %s\r\n", encodeBasicAuth(secret))
	}
	chunked := len(r.TransferEncoding) > 0 && r.TransferEncoding[0] == "chunked"
	if len(r.TransferEncoding) > 0 {
		fmt.Fprintf(&b, "Transfer-Encoding: %s\r\n", strings.Join(r.TransferEncoding, ","))
	}
	if r.Close {
		fmt.Fprintf(&b, "Connection: close\r\n")
	}

	err = r.Header.WriteSubset(&b, nil)
	if err != nil {
		return nil, err
	}

	io.WriteString(&b, "\r\n")

	if r.Body != nil {
		var dest io.Writer = &b
		if chunked {
			dest = httputil.NewChunkedWriter(dest)
		}
		_, err = io.Copy(dest, r.Body)
		if chunked {
			dest.(io.Closer).Close()
			io.WriteString(&b, "\r\n")
		}
	}

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (t *HttpProxy) isMatch(value string, rules []string) bool {
	if len(rules) == 0 {
		return true
	}
	for _, rule := range rules {
		pattern := fmt.Sprintf(`(^|\.)%s(:\d+)?$`, regexp.QuoteMeta(rule))
		re := regexp.MustCompile(pattern)
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func (t *HttpProxy) findForward(r *http.Request) *co.HttpForward {
	t.lock.RLock()
	defer t.lock.RUnlock()

	via := t.cfg.Forward
	if via != nil && t.isMatch(r.URL.Host, via.Match) {
		return via
	}
	for _, via := range t.cfg.Backends {
		if via != nil && t.isMatch(r.URL.Host, via.Match) {
			return via
		}
	}
	return nil
}

func (t *HttpProxy) doRecord(r *http.Request) {
	t.lock.Lock()
	defer t.lock.Unlock()

	record, ok := t.requests[r.URL.Host]
	if !ok {
		record = &HttpRecord{}
		t.requests[r.URL.Host] = record
	}
	record.Update()
}

func (t *HttpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.out.Debug("HttpProxy.ServeHTTP %v", r.URL)

	if !t.CheckAuth(w, r) {
		t.out.Info("HttpProxy.ServeHTTP Required %v Authentication", r.URL)
		return
	}
	if r.URL.Host == "" {
		t.api.ServeHTTP(w, r)
		return
	}

	t.doRecord(r)
	via := t.findForward(r)
	if via != nil {
		t.out.Info("HttpProxy.ServeHTTP %s %s -> %s via %s", r.Method, r.RemoteAddr, r.URL.Host, via.Server)
		conn, err := t.openConn(via.Protocol, via.Server, via.Insecure)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		dump, err := t.cloneRequest(r, via.Secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		conn.Write(dump)
		t.toTunnel(w, conn)
		return
	}

	t.out.Info("HttpProxy.ServeHTTP %s %s -> %s", r.Method, r.RemoteAddr, r.URL.Host)
	if r.Method == "CONNECT" { //RFC-7231 Tunneling TCP based protocols through Web Proxy servers
		conn, err := t.openConn("", r.URL.Host, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Write(connectOkay)
		t.toTunnel(w, conn)
	} else { //RFC 7230 - HTTP/1.1: Message Syntax and Routing
		transport := &http.Transport{}
		p, err := transport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer transport.CloseIdleConnections()
		t.toDirect(w, p)
	}
}

func (t *HttpProxy) Start() {
	if t.server == nil || t.cfg == nil {
		return
	}
	crt := t.cfg.Cert
	if crt == nil || crt.KeyFile == "" {
		t.out.Info("HttpProxy.start http://%s", t.server.Addr)
	} else {
		t.out.Info("HttpProxy.start https://%s", t.server.Addr)
	}
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		t.startat = time.Now()
		if crt == nil || crt.KeyFile == "" {
			if err := t.server.ListenAndServe(); err != nil {
				t.out.Warn("HttpProxy.start %s", err)
				return err
			}
		} else {
			if err := t.server.ListenAndServeTLS(crt.CrtFile, crt.KeyFile); err != nil {
				t.out.Error("HttpProxy.start %s", err)
				return err
			}
		}
		t.server.Shutdown(nil)
		return nil
	})
}

var httpTmpl = map[string]string{
	"stats": `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>OpenLAN Proxy</title>
    <style>
      body, table, a {
        background-color: #1d1d1d;
        color: #bababa;
        font-size: small;
      }
      td, th {
        border: 1px solid #424242;
        border-radius: 2px;
        text-align: center;
      }
    </style>
  </head>
  <body>
    <table>
    <tr>
      <td>Total:</td><td>{{ .Total }}</td>
    </tr>
    <tr>
      <td>Configuration:</td><td><a href="/api/config">display</a></td>
    </tr>
    <tr>
      <td>APIs:</td><td><a href="/api">display</a></td>
    </tr>
    <tr>
      <td>StartAt:</td><td>{{ .StartAt }}</td>
    </tr>
    </table>
    <table>
    <tr>
      <td>Domain</td><td>Count</td><td>LastAt</td><td>CreateAt</td>
    </tr>
    {{- range .Requests }}
    <tr>
      <td>{{ .Domain }}</td><td>{{ .Count }}</td>
      <td>{{ .LastAt }}</td><td>{{ .CreateAt }}</td>
    </tr>
    {{- end }}
    </table>
  <body>
</html>`,
	"pac": `
function FindProxyForURL(url, host) {
    if (isPlainHostName(host))
        return "DIRECT";
{{- range .Rules }}
    if (shExpMatch(host, "(^|*\.){{ . }}"))
        return "PROXY {{ $.Local }}";
{{- end }}
    return "DIRECT";
}`,
}

func (t *HttpProxy) GetStats(w http.ResponseWriter, r *http.Request) {
	t.lock.RLock()
	data := &struct {
		StartAt  string
		Total    int
		Requests []*HttpRecord
	}{
		Total:   len(t.requests),
		StartAt: t.startat.Local().String(),
	}
	for name, record := range t.requests {
		data.Requests = append(data.Requests, &HttpRecord{
			Domain:   name,
			Count:    record.Count,
			LastAt:   record.LastAt,
			CreateAt: record.CreateAt,
		})
	}
	t.lock.RUnlock()

	sort.SliceStable(data.Requests, func(i, j int) bool {
		ii := data.Requests[i]
		jj := data.Requests[j]
		return ii.LastAt > jj.LastAt
	})

	if t.findQuery(r, "format") == "yaml" {
		encodeYaml(w, data)
	} else if t.findQuery(r, "format") == "json" {
		encodeJson(w, data)
	} else {
		encodeText(w, httpTmpl["stats"], data)
	}
}

func (t *HttpProxy) GetConfig(w http.ResponseWriter, r *http.Request) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.findQuery(r, "format") == "json" {
		encodeJson(w, t.cfg)
	} else {
		encodeYaml(w, t.cfg)
	}
}

func (t *HttpProxy) AddMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	domain := vars["domain"]
	backend := vars["backend"]

	t.lock.Lock()
	defer t.lock.Unlock()

	if t.cfg.AddMatch(domain, backend) > -1 {
		encodeYaml(w, "success")
	} else {
		encodeYaml(w, "failed")
	}
	t.Save()
}

func (t *HttpProxy) Save() {
	if t.proxer == nil {
		t.cfg.Save()
	} else {
		t.proxer.Save()
	}
}

func (t *HttpProxy) DelMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	domain := vars["domain"]
	backend := vars["backend"]

	t.lock.Lock()
	defer t.lock.Unlock()

	if t.cfg.DelMatch(domain, backend) > -1 {
		encodeYaml(w, "success")
	} else {
		encodeYaml(w, "failed")
	}
	t.Save()
}

func (t *HttpProxy) GetPac(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Local string
		Rules []string
	}{
		Local: t.cfg.Listen,
	}

	for _, via := range t.cfg.Backends {
		for _, rule := range via.Match {
			data.Rules = append(data.Rules, rule)
		}
	}

	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	encodeText(w, httpTmpl["pac"], data)
}

func (t *HttpProxy) findQuery(r *http.Request, name string) string {
	query := r.URL.Query()
	if values, ok := query[name]; ok {
		return values[0]
	}
	return ""
}

func (t *HttpProxy) GetApi(w http.ResponseWriter, r *http.Request) {
	var urls []string

	t.api.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		for _, m := range methods {
			urls = append(urls, fmt.Sprintf("%-6s %s", m, path))
		}
		return nil
	})
	encodeYaml(w, urls)
}
