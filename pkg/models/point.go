package models

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
)

type Point struct {
	UUID     string             `json:"uuid"`
	Alias    string             `json:"alias"`
	Network  string             `json:"network"`
	User     string             `json:"user"`
	Protocol string             `json:"protocol"`
	Server   string             `json:"server"`
	Uptime   int64              `json:"uptime"`
	Status   string             `json:"status"`
	IfName   string             `json:"device"`
	Client   libol.SocketClient `json:"-"`
	Device   network.Taper      `json:"-"`
	System   string             `json:"system"`
}

func NewPoint(c libol.SocketClient, d network.Taper, proto string) (w *Point) {
	return &Point{
		Alias:    "",
		Server:   c.LocalAddr(),
		Client:   c,
		Device:   d,
		Protocol: proto,
	}
}

func (p *Point) Update() *Point {
	client := p.Client
	if client != nil {
		p.Uptime = client.UpTime()
		p.Status = client.Status().String()
	}
	device := p.Device
	if device != nil {
		p.IfName = device.Name()
	}
	return p
}

func (p *Point) SetUser(user *User) {
	p.User = user.Name
	p.UUID = user.UUID
	if len(p.UUID) > 13 {
		// too long and using short uuid.
		p.UUID = p.UUID[:13]
	}
	p.Network = user.Network
	p.System = user.System
	p.Alias = user.Alias
}
