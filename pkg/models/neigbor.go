package models

import (
	"net"
	"time"

	"github.com/luscis/openlan/pkg/libol"
)

type Neighbor struct {
	Network string           `json:"network"`
	Device  string           `json:"device"`
	Client  string           `json:"client"`
	HwAddr  net.HardwareAddr `json:"hwAddr"`
	IpAddr  net.IP           `json:"ipAddr"`
	NewTime int64            `json:"newTime"`
	HitTime int64            `json:"hitTime"`
}

func (e *Neighbor) String() string {
	str := e.HwAddr.String()
	str += ":" + e.IpAddr.String()
	str += ":" + e.Client
	return str
}

func NewNeighbor(hwAddr net.HardwareAddr, ipAddr net.IP, client libol.SocketClient) (e *Neighbor) {
	e = &Neighbor{
		HwAddr:  hwAddr,
		IpAddr:  ipAddr,
		Client:  client.String(),
		NewTime: time.Now().Unix(),
		HitTime: time.Now().Unix(),
	}
	e.Update(client)
	return
}

func (e *Neighbor) UpTime() int64 {
	return time.Now().Unix() - e.HitTime
}

func (e *Neighbor) Update(client libol.SocketClient) {
	if client == nil {
		return
	}
	private := client.Private()
	if private == nil {
		return
	}
	if acc, ok := private.(*Access); ok {
		e.Network = acc.Network
		e.Device = acc.IfName
		e.Client = client.String()
	}
}
