package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type output struct {
	outputs *libol.SafeStrMap
}

func (p *output) Init(size int) {
	p.outputs = libol.NewSafeStrMap(size)
}

func (p *output) Add(uuid string, output *models.Output) {
	_ = p.outputs.Set(uuid, output)
}

func (p *output) Get(key string) *models.Output {
	ret := p.outputs.Get(key)
	if ret != nil {
		return ret.(*models.Output)
	}
	return nil
}

func (p *output) Del(key string) {
	p.outputs.Del(key)
}

func (p *output) List(name string) <-chan *models.Output {
	c := make(chan *models.Output, 128)
	go func() {
		p.outputs.Iter(func(k string, v interface{}) {
			m := v.(*models.Output)
			if name == "" || m.Network == name {
				m.Update()
				c <- m
			}
		})
		c <- nil //Finish channel by nil.
	}()
	return c
}

func (p *output) ListAll() <-chan *models.Output {
	return p.List("")
}

var Output = output{
	outputs: libol.NewSafeStrMap(1024),
}
