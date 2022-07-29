package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type online struct {
	Lines *libol.SafeStrMap
}

func (p *online) Init(size int) {
	p.Lines = libol.NewSafeStrMap(size)
}

func (p *online) Add(m *models.Line) {
	_ = p.Lines.Set(m.String(), m)
}

func (p *online) Update(m *models.Line) *models.Line {
	if v := p.Lines.Get(m.String()); v != nil {
		l := v.(*models.Line)
		l.HitTime = m.HitTime
	}
	return nil
}

func (p *online) Get(key string) *models.Line {
	if v := p.Lines.Get(key); v != nil {
		return v.(*models.Line)
	}
	return nil
}

func (p *online) Del(key string) {
	p.Lines.Del(key)
}

func (p *online) List() <-chan *models.Line {
	c := make(chan *models.Line, 128)

	go func() {

		p.Lines.Iter(func(k string, v interface{}) {
			c <- v.(*models.Line)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

var Online = online{
	Lines: libol.NewSafeStrMap(1024),
}
