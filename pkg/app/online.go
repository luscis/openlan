package app

import (
	"container/list"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type Online struct {
	lock     sync.RWMutex
	maxSize  int
	lineMap  map[string]*models.Line
	lineList *list.List
	master   Master
}

func NewOnline(m Master) *Online {
	c := config.Get()
	ms := c.Limit.OnLine
	return &Online{
		maxSize:  ms,
		lineMap:  make(map[string]*models.Line, ms),
		lineList: list.New(),
		master:   m,
	}
}

func (o *Online) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if frame.IsControl() {
		return nil
	}
	if libol.HasLog(libol.LOG) {
		libol.Log("Online.OnFrame %s.", frame)
	}
	proto, err := frame.Proto()
	if err != nil {
		libol.Warn("Online.OnFrame %s", err)
		return err
	}
	if proto.Ip4 != nil {
		ip := proto.Ip4
		line := models.NewLine(libol.EthIp4)
		line.IpSource = ip.Source
		line.IpDest = ip.Destination
		line.IpProtocol = ip.Protocol
		if proto.Tcp != nil {
			tcp := proto.Tcp
			line.PortDest = tcp.Destination
			line.PortSource = tcp.Source
		} else if proto.Udp != nil {
			udp := proto.Udp
			line.PortDest = udp.Destination
			line.PortSource = udp.Source
		}
		o.AddLine(line)
	}
	return nil
}

func (o *Online) popLine() {
	if o.lineList.Len() < o.maxSize {
		return
	}
	e := o.lineList.Front()
	if e == nil {
		return
	}
	if lastLine, ok := e.Value.(*models.Line); ok {
		o.lineList.Remove(e)
		cache.Online.Del(lastLine.String())
		delete(o.lineMap, lastLine.String())
	}
}

func (o *Online) AddLine(line *models.Line) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if libol.HasLog(libol.LOG) {
		libol.Log("Online.AddLine %s and len %d", line, o.lineList.Len())
	}
	key := line.String()
	if find, ok := o.lineMap[key]; !ok {
		o.popLine()
		o.lineList.PushBack(line)
		o.lineMap[key] = line
		cache.Online.Add(line)
	} else if find != nil {
		find.HitTime = time.Now().Unix()
		cache.Online.Update(find)
	}
}
