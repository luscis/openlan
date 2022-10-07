package main

import (
	"flag"
	"fmt"
	db "github.com/luscis/openlan/pkg/database"
	"github.com/luscis/openlan/pkg/libol"
	"time"
)

type Config struct {
	UdpPort  int
	LogLevel int
	LogFile  string
}

func (c *Config) Parse() {
	flag.IntVar(&c.UdpPort, "port", 4500, "UDP port listen on")
	flag.StringVar(&c.LogFile, "log:file", "/dev/null", "File log saved to")
	flag.IntVar(&c.LogLevel, "log:level", 20, "Log level value")
	flag.Parse()
}

type UdpServer struct {
	stop   chan struct{}
	out    *libol.SubLogger
	server *libol.UdpInServer
	cfg    *Config
	links  *libol.SafeStrMap
}

func NewUdpServer(cfg *Config) *UdpServer {
	c := &UdpServer{
		out:   libol.NewSubLogger("udp"),
		stop:  make(chan struct{}),
		cfg:   cfg,
		links: libol.NewSafeStrMap(128),
	}
	return c
}

func (u *UdpServer) Initialize() {
	u.server = &libol.UdpInServer{
		Port: uint16(u.cfg.UdpPort),
	}
}

func (u *UdpServer) Start() {
	if err := u.server.Open(); err != nil {
		u.out.Error("UdpServer.Start open socket %s", err)
		return
	}
}

func (u *UdpServer) Stop() {
}

func (u *UdpServer) Device2UUID(value string) string {
	if link := u.links.Get(value); link != nil {
		if older, ok := link.(*db.VirtualLink); ok {
			return older.UUID
		}
	}
	return ""
}

func (u *UdpServer) toStatus(li *db.DBClient, from *libol.UdpInConnection) {
	device := fmt.Sprintf("spi:%d", from.Spi)
	obj := &db.VirtualLink{
		UUID: u.Device2UUID(device),
	}
	if err := li.Client.Get(obj); err != nil {
		return
	}
	if obj.Status == nil {
		obj.Status = make(map[string]string, 2)
	}
	obj.Status["remote_connection"] = fmt.Sprintf("udp:%s", from.Connection())
	ops, err := li.Client.Where(obj).Update(obj)
	if err != nil {
		u.out.Warn("UdpServer.toStatus update %s", err)
		return
	}
	if _, err := li.Client.Transact(ops...); err != nil {
		u.out.Warn("UdpServer.toStatus commit %s", err)
		return
	}
	if obj.Connection == "any" {
		_ = u.server.Send(from)
	}
}

func (u *UdpServer) toLinkState(li *db.DBClient, from *libol.UdpInConnection) {
	device := fmt.Sprintf("spi:%d", from.Spi)
	obj := &db.VirtualLink{
		UUID: u.Device2UUID(device),
	}
	if err := li.Client.Get(obj); err != nil {
		return
	}
	obj.LinkState = "up"
	ops, err := li.Client.Where(obj).Update(obj)
	if err != nil {
		u.out.Warn("UdpServer.toLinkState update %s", err)
		return
	}
	if _, err := li.Client.Transact(ops...); err != nil {
		u.out.Warn("UdpServer.toLinkState commit %s", err)
		return
	}
}

func (u *UdpServer) Pong() {
	li, err := db.NewClient(nil)
	if err != nil {
		u.out.Error("UdpServer.Pong open db with %s", err)
		return
	}

	for {
		from, _ := u.server.Recv()
		u.out.Cmd("UdpServer.Pong received %s", from.String())

		u.toStatus(li, from)
		u.toLinkState(li, from)
	}
}

func (u *UdpServer) toPing(li *db.DBClient, obj *db.VirtualLink) {
	addr, port := db.GetAddrPort(obj.Connection[4:])
	if port == 0 {
		port = 4500
	}
	conn := &libol.UdpInConnection{
		Spi:        obj.Spi(),
		RemotePort: uint16(port),
		RemoteAddr: addr,
	}
	u.out.Cmd("UdpServer.toPing send to %s", conn.String())
	_ = u.server.Send(conn)
}

func (u *UdpServer) Ping() {
	li, err := db.NewClient(nil)
	if err != nil {
		u.out.Error("UdpServer.Ping open db with %s", err)
		return
	}

	for {
		var ls []db.VirtualLink
		_ = li.Client.List(&ls)
		u.links.Clear()
		for i := range ls {
			obj := &ls[i]
			if err := u.links.Mod(obj.Device, obj); err != nil {
				u.out.Error("UdpServer.Ping %s", err)
			}
			if !obj.IsUdpIn() {
				continue
			}
			u.toPing(li, obj)
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	c := &Config{}
	c.Parse()

	libol.SetLogger(c.LogFile, c.LogLevel)

	srv := NewUdpServer(c)
	srv.Initialize()

	srv.Start()
	libol.Go(srv.Ping)
	libol.Go(srv.Pong)

	libol.Wait()
	srv.Stop()
}
