package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type neighbor struct {
	Neighbors *libol.SafeStrMap
}

func (p *neighbor) Init(size int) {
	p.Neighbors = libol.NewSafeStrMap(size)
}

func (p *neighbor) Add(m *models.Neighbor) {
	if v := p.Neighbors.Get(m.IpAddr.String()); v != nil {
		p.Neighbors.Del(m.IpAddr.String())
	}
	_ = p.Neighbors.Set(m.IpAddr.String(), m)
}

func (p *neighbor) Update(m *models.Neighbor) *models.Neighbor {
	if v := p.Neighbors.Get(m.IpAddr.String()); v != nil {
		n := v.(*models.Neighbor)
		n.HwAddr = m.HwAddr
		n.HitTime = m.HitTime
	}
	return nil
}

func (p *neighbor) Get(key string) *models.Neighbor {
	if v := p.Neighbors.Get(key); v != nil {
		return v.(*models.Neighbor)
	}
	return nil
}

func (p *neighbor) Del(key string) {
	p.Neighbors.Del(key)
}

func (p *neighbor) List() <-chan *models.Neighbor {
	c := make(chan *models.Neighbor, 128)

	go func() {
		p.Neighbors.Iter(func(k string, v interface{}) {
			c <- v.(*models.Neighbor)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

var Neighbor = neighbor{
	Neighbors: libol.NewSafeStrMap(1024),
}
