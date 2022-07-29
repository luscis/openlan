package app

import (
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"net"
	"time"
)

type Neighbors struct {
	master Master
}

func NewNeighbors(m Master) *Neighbors {
	return &Neighbors{
		master: m,
	}
}

func (e *Neighbors) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if frame.IsControl() {
		return nil
	}
	if libol.HasLog(libol.LOG) {
		libol.Log("Neighbors.OnFrame %s.", frame)
	}
	proto, err := frame.Proto()
	if err != nil {
		libol.Warn("Neighbors.OnFrame %s", err)
		return err
	}
	if eth := proto.Eth; !eth.IsArp() {
		return nil
	}
	arp := proto.Arp
	if arp.IsIP4() && (arp.IsReply() || arp.IsRequest()) {
		n := models.NewNeighbor(arp.SHwAddr, arp.SIpAddr, client)
		e.AddNeighbor(n, client)
	}
	return nil
}

func (e *Neighbors) AddNeighbor(new *models.Neighbor, client libol.SocketClient) {
	if n := cache.Neighbor.Get(new.IpAddr.String()); n != nil {
		libol.Log("Neighbors.AddNeighbor: update %s.", new)
		n.Update(client)
		n.HitTime = time.Now().Unix()
	} else {
		libol.Log("Neighbors.AddNeighbor: new %s.", new)
		cache.Neighbor.Add(new)
	}
}

func (e *Neighbors) DelNeighbor(ipAddr net.IP) {
	libol.Info("Neighbors.DelNeighbor %s.", ipAddr)
	if n := cache.Neighbor.Get(ipAddr.String()); n != nil {
		cache.Neighbor.Del(ipAddr.String())
	}
}

func (e *Neighbors) OnClientClose(client libol.SocketClient) {
	//TODO
	libol.Info("Neighbors.OnClientClose %s.", client)
}
