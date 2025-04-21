package cache

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type access struct {
	Clients  *libol.SafeStrMap
	UUIDAddr *libol.SafeStrStr
	AddrUUID *libol.SafeStrStr
}

func (p *access) Init(size int) {
	p.Clients = libol.NewSafeStrMap(size)
	p.UUIDAddr = libol.NewSafeStrStr(size)
	p.AddrUUID = libol.NewSafeStrStr(size)
}

func (p *access) Add(m *models.Access) {
	_ = p.UUIDAddr.Reset(m.UUID, m.Client.String())
	_ = p.AddrUUID.Set(m.Client.String(), m.UUID)
	_ = p.Clients.Set(m.Client.String(), m)
}

func (p *access) Get(addr string) *models.Access {
	if v := p.Clients.Get(addr); v != nil {
		m := v.(*models.Access)
		m.Update()
		return m
	}
	return nil
}

func (p *access) GetByUUID(uuid string) *models.Access {
	if addr := p.GetAddr(uuid); addr != "" {
		return p.Get(addr)
	}
	return nil
}

func (p *access) GetUUID(addr string) string {
	return p.AddrUUID.Get(addr)
}

func (p *access) GetAddr(uuid string) string {
	return p.UUIDAddr.Get(uuid)
}

func (p *access) Del(addr string) {
	if v := p.Clients.Get(addr); v != nil {
		m := v.(*models.Access)
		if m.Device != nil {
			_ = m.Device.Close()
		}
		if p.UUIDAddr.Get(m.UUID) == addr { // not has newer
			p.UUIDAddr.Del(m.UUID)
		}
		p.AddrUUID.Del(m.Client.String())
		p.Clients.Del(addr)
	}
}

func (p *access) List() <-chan *models.Access {
	c := make(chan *models.Access, 128)

	go func() {
		p.Clients.Iter(func(k string, v interface{}) {
			if m, ok := v.(*models.Access); ok {
				m.Update()
				c <- m
			}
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

var Access = access{
	Clients:  libol.NewSafeStrMap(1024),
	UUIDAddr: libol.NewSafeStrStr(1024),
	AddrUUID: libol.NewSafeStrStr(1024),
}
