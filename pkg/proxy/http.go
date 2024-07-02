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
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
)

type HttpProxy struct {
	users    map[string]string
	out      *libol.SubLogger
	server   *http.Server
	cfg      *config.HttpProxy
	api      *mux.Router
	requests int
	startat  time.Time
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

func NewHttpProxy(cfg *config.HttpProxy) *HttpProxy {
	h := &HttpProxy{
		out:   libol.NewSubLogger(cfg.Listen),
		cfg:   cfg,
		users: make(map[string]string),
		api:   mux.NewRouter(),
	}

	h.server = &http.Server{
		Addr:    cfg.Listen,
		Handler: h,
	}
	auth := cfg.Auth
	if auth != nil && auth.Username != "" {
		h.users[auth.Username] = auth.Password
	}

	h.loadUrl()
	h.loadPass()

	return h
}

func (t *HttpProxy) loadUrl() {
	if strings.HasPrefix(t.cfg.Listen, "127.") {
		t.api.HandleFunc("/", t.GetStats)
		t.api.HandleFunc("/config", t.GetConfig)
	}
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
		t.users[user] = pass
	}
	if err := scanner.Err(); err != nil {
		libol.Warn("HttpProxy.LoadPass scaner %v", err)
	}
}

func (t *HttpProxy) isAuth(username, password string) bool {
	if p, ok := t.users[username]; ok {
		return p == password
	}
	return false
}

func (t *HttpProxy) CheckAuth(w http.ResponseWriter, r *http.Request) bool {
	if len(t.users) == 0 {
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

func (t *HttpProxy) route(w http.ResponseWriter, p *http.Response) {
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

func (t *HttpProxy) tunnel(w http.ResponseWriter, conn net.Conn) {
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
		return tls.Dial("tcp", remote, conf)
	}
	return net.Dial("tcp", remote)
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

func (t *HttpProxy) Match(value string, rules []string) bool {
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

	t.requests += 1
	fow := t.cfg.Forward
	if fow != nil && t.Match(r.URL.Host, fow.Match) {
		t.out.Info("HttpProxy.ServeHTTP %s %s -> %s via %s", r.Method, r.RemoteAddr, r.URL.Host, fow.Server)
		conn, err := t.openConn(fow.Protocol, fow.Server, fow.Insecure)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		dump, err := t.cloneRequest(r, fow.Secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		conn.Write(dump)
		t.tunnel(w, conn)
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
		t.tunnel(w, conn)
	} else { //RFC 7230 - HTTP/1.1: Message Syntax and Routing
		transport := &http.Transport{}
		p, err := transport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer transport.CloseIdleConnections()
		t.route(w, p)
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

func (t *HttpProxy) GetStats(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Requests int
		StartAt  time.Time
	}{
		Requests: t.requests,
		StartAt:  t.startat,
	}
	encodeJson(w, data)
}

func (t *HttpProxy) GetConfig(w http.ResponseWriter, r *http.Request) {
	encodeJson(w, t.cfg)
}
