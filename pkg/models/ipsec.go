package models

import (
	"fmt"
	"time"

	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

type Esp struct {
	Name    string
	Address string
	NewTime int64
}

func (l *Esp) Update() {
}

func (l *Esp) ID() string {
	return l.Name
}

func NewEspSchema(e *Esp) schema.Esp {
	e.Update()
	se := schema.Esp{
		Name:    e.Name,
		Address: e.Address,
	}
	return se
}

type EspState struct {
	*schema.EspState
	NewTime int64
	In      *nl.XfrmState
	Out     *nl.XfrmState
}

func (l *EspState) ID() string {
	return fmt.Sprintf("spi:%d %s-%s", l.Spi, l.Local, l.Remote)
}

func (l *EspState) UpTime() int64 {
	return time.Now().Unix() - l.NewTime
}

func (l *EspState) String() string {
	return fmt.Sprintf("{Spi: %d Local: %s Remote: %s}", l.Spi, l.Local, l.Remote)
}

func NewEspStateSchema(e *EspState) schema.EspState {
	e.Update()
	se := schema.EspState{
		Name:       e.Name,
		Spi:        e.Spi,
		Local:      e.Local,
		Remote:     e.Remote,
		TxBytes:    e.TxBytes,
		TxPackages: e.TxPackages,
		RxBytes:    e.RxBytes,
		RxPackages: e.RxPackages,
		AliveTime:  e.AliveTime,
	}
	return se
}

type EspPolicy struct {
	*schema.EspPolicy
	In  *nl.XfrmPolicy
	Fwd *nl.XfrmPolicy
	Out *nl.XfrmPolicy
}

func (l *EspPolicy) ID() string {
	return fmt.Sprintf("spi:%d %s-%s", l.Spi, l.Source, l.Dest)
}

func (l *EspPolicy) String() string {
	return fmt.Sprintf("{Spi: %d Source: %s Dest: %s}", l.Spi, l.Source, l.Dest)
}

func (l *EspPolicy) Update() {
}

func NewEspPolicySchema(e *EspPolicy) schema.EspPolicy {
	e.Update()
	se := schema.EspPolicy{
		Name:   e.Name,
		Source: e.Source,
		Dest:   e.Dest,
	}
	return se
}
