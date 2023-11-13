package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type EspSmap struct {
	Esp *libol.SafeStrMap
}

func (p *EspSmap) Init(size int) {
	p.Esp = libol.NewSafeStrMap(size)
}

func (p *EspSmap) Add(EspSmap *models.Esp) {
	_ = p.Esp.Set(EspSmap.ID(), EspSmap)
}

func (p *EspSmap) Get(key string) *models.Esp {
	ret := p.Esp.Get(key)
	if ret != nil {
		return ret.(*models.Esp)
	}
	return nil
}

func (p *EspSmap) Del(key string) {
	p.Esp.Del(key)
}

func (p *EspSmap) List() <-chan *models.Esp {
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

var Esp = EspSmap{
	Esp: libol.NewSafeStrMap(1024),
}

type EspSmapState struct {
	State *libol.SafeStrMap
}

func (p *EspSmapState) Init(size int) {
	p.State = libol.NewSafeStrMap(size)
}

func (p *EspSmapState) Add(EspSmap *models.EspState) {
	_ = p.State.Set(EspSmap.ID(), EspSmap)
}

func (p *EspSmapState) Get(key string) *models.EspState {
	ret := p.State.Get(key)
	if ret != nil {
		return ret.(*models.EspState)
	}
	return nil
}

func (p *EspSmapState) Del(key string) {
	p.State.Del(key)
}

func (p *EspSmapState) List(name string) <-chan *models.EspState {
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

func (p *EspSmapState) Clear() {
	p.State.Clear()
}

var EspState = EspSmapState{
	State: libol.NewSafeStrMap(1024),
}

type EspSmapPolicy struct {
	Policy *libol.SafeStrMap
}

func (p *EspSmapPolicy) Init(size int) {
	p.Policy = libol.NewSafeStrMap(size)
}

func (p *EspSmapPolicy) Add(EspSmap *models.EspPolicy) {
	_ = p.Policy.Set(EspSmap.ID(), EspSmap)
}

func (p *EspSmapPolicy) Get(key string) *models.EspPolicy {
	ret := p.Policy.Get(key)
	if ret != nil {
		return ret.(*models.EspPolicy)
	}
	return nil
}

func (p *EspSmapPolicy) Del(key string) {
	p.Policy.Del(key)
}

func (p *EspSmapPolicy) List(name string) <-chan *models.EspPolicy {
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

func (p *EspSmapPolicy) Clear() {
	p.Policy.Clear()
}

var EspPolicy = EspSmapPolicy{
	Policy: libol.NewSafeStrMap(1024),
}
