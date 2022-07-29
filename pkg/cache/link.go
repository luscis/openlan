package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type link struct {
	Links *libol.SafeStrMap
}

func (p *link) Init(size int) {
	p.Links = libol.NewSafeStrMap(size)
}

func (p *link) Add(uuid string, link *models.Link) {
	_ = p.Links.Set(uuid, link)
}

func (p *link) Get(key string) *models.Link {
	ret := p.Links.Get(key)
	if ret != nil {
		return ret.(*models.Link)
	}
	return nil
}

func (p *link) Del(key string) {
	p.Links.Del(key)
}

func (p *link) List() <-chan *models.Link {
	c := make(chan *models.Link, 128)
	go func() {
		p.Links.Iter(func(k string, v interface{}) {
			m := v.(*models.Link)
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()
	return c
}

var Link = link{
	Links: libol.NewSafeStrMap(1024),
}
