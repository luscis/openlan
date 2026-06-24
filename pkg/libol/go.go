package libol

import (
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
)

type gos struct {
	lock  sync.Mutex
	total uint64
}

var Gos = gos{}

func (t *gos) Add(call any) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.total++
	Debug("gos.Add %d %p", t.total, call)
}

func (t *gos) Del(call any) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.total--
	Debug("gos.Del %d %p", t.total, call)
}

func Go(call func()) {
	name := FunName(call)
	go func() {
		defer Catch("Go.func")
		Gos.Add(call)
		Debug("Go.Add: %s", name)
		call()
		Debug("Go.Del: %s", name)
		Gos.Del(call)
	}()
}

type PProf struct {
	File   string
	Listen string
	Error  error
}

type pprofHandler struct{}

func (pprofHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		pprof.Index(w, r)
		return
	case "/cmdline":
		pprof.Cmdline(w, r)
		return
	case "/profile":
		pprof.Profile(w, r)
		return
	case "/symbol":
		pprof.Symbol(w, r)
		return
	case "/trace":
		pprof.Trace(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/")
	pprof.Handler(name).ServeHTTP(w, r)
}

func (p *PProf) Start() {
	if p.Listen == "" {
		p.Listen = "localhost:6060"
	}
	Go(func() {
		Info("PProf.Start %s", p.Listen)
		if err := http.ListenAndServe(p.Listen, pprofHandler{}); err != nil {
			Error("PProf.Start %s", err)
			p.Error = err
		}
	})
}

func (p *PProf) Stop() {
}
