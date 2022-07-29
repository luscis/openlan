package libol

import (
	"net/http"
	_ "net/http/pprof"
	"sync"
)

type gos struct {
	lock  sync.Mutex
	total uint64
}

var Gos = gos{}

func (t *gos) Add(call interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.total++
	Debug("gos.Add %d %p", t.total, call)
}

func (t *gos) Del(call interface{}) {
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

func (p *PProf) Start() {
	if p.Listen == "" {
		p.Listen = "localhost:6060"
	}
	Go(func() {
		Info("PProf.Start %s", p.Listen)
		if err := http.ListenAndServe(p.Listen, nil); err != nil {
			Error("PProf.Start %s", err)
			p.Error = err
		}
	})
}

func (p *PProf) Stop() {
}
