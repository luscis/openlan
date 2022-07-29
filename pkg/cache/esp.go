package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type esp struct {
	Esp *libol.SafeStrMap
}

func (p *esp) Init(size int) {
	p.Esp = libol.NewSafeStrMap(size)
}

func (p *esp) Add(esp *models.Esp) {
	_ = p.Esp.Set(esp.ID(), esp)
}

func (p *esp) Get(key string) *models.Esp {
	ret := p.Esp.Get(key)
	if ret != nil {
		return ret.(*models.Esp)
	}
	return nil
}

func (p *esp) Del(key string) {
	p.Esp.Del(key)
}

func (p *esp) List() <-chan *models.Esp {
	c := make(chan *models.Esp, 128)
	go func() {
		p.Esp.Iter(func(k string, v interface{}) {
			m := v.(*models.Esp)
			m.Update()
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()
	return c
}

var Esp = esp{
	Esp: libol.NewSafeStrMap(1024),
}

type espState struct {
	State *libol.SafeStrMap
}

func (p *espState) Init(size int) {
	p.State = libol.NewSafeStrMap(size)
}

func (p *espState) Add(esp *models.EspState) {
	_ = p.State.Set(esp.ID(), esp)
}

func (p *espState) Get(key string) *models.EspState {
	ret := p.State.Get(key)
	if ret != nil {
		return ret.(*models.EspState)
	}
	return nil
}

func (p *espState) Del(key string) {
	p.State.Del(key)
}

func (p *espState) List(name string) <-chan *models.EspState {
	c := make(chan *models.EspState, 128)
	go func() {
		p.State.Iter(func(k string, v interface{}) {
			m := v.(*models.EspState)
			if m.Name == name || name == "" {
				m.Update()
				c <- m
			}
		})
		c <- nil //Finish channel by nil.
	}()
	return c
}

func (p *espState) Clear() {
	p.State.Clear()
}

var EspState = espState{
	State: libol.NewSafeStrMap(1024),
}

type espPolicy struct {
	Policy *libol.SafeStrMap
}

func (p *espPolicy) Init(size int) {
	p.Policy = libol.NewSafeStrMap(size)
}

func (p *espPolicy) Add(esp *models.EspPolicy) {
	_ = p.Policy.Set(esp.ID(), esp)
}

func (p *espPolicy) Get(key string) *models.EspPolicy {
	ret := p.Policy.Get(key)
	if ret != nil {
		return ret.(*models.EspPolicy)
	}
	return nil
}

func (p *espPolicy) Del(key string) {
	p.Policy.Del(key)
}

func (p *espPolicy) List(name string) <-chan *models.EspPolicy {
	c := make(chan *models.EspPolicy, 128)
	go func() {
		p.Policy.Iter(func(k string, v interface{}) {
			m := v.(*models.EspPolicy)
			if m.Name == name || name == "" {
				m.Update()
				c <- m
			}
		})
		c <- nil //Finish channel by nil.
	}()
	return c
}

func (p *espPolicy) Clear() {
	p.Policy.Clear()
}

var EspPolicy = espPolicy{
	Policy: libol.NewSafeStrMap(1024),
}
