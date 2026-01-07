package cswitch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	nl "github.com/vishvananda/netlink"
)

const (
	AccessBin = "openlan-access"
	AccessDir = "/var/openlan/access"
)

type Link struct {
	cfg  *co.Access
	out  *libol.SubLogger
	uuid string
}

func NewLink(cfg *co.Access) *Link {
	uuid := libol.GenString(13)
	return &Link{
		uuid: uuid,
		cfg:  cfg,
		out:  libol.NewSubLogger(cfg.Network),
	}
}

func (l *Link) Initialize() {
	file := l.ConfFile()
	l.cfg.Log = co.Log{File: l.LogFile()}
	l.cfg.StatusFile = l.StatusFile()
	l.cfg.PidFile = l.PidFile()
	_ = libol.MarshalSave(l.cfg, file, true)
}

func (l *Link) Conf() *co.Access {
	return l.cfg
}

func (l *Link) UUID() string {
	return l.uuid
}

func (l *Link) Path() string {
	return AccessBin
}

func (l *Link) ID() string {
	return l.cfg.ID()
}

func (l *Link) ConfFile() string {
	return filepath.Join(AccessDir, l.ID()+".json")
}

func (l *Link) StatusFile() string {
	return filepath.Join(AccessDir, l.ID()+".status")
}

func (l *Link) PidFile() string {
	return filepath.Join(AccessDir, l.ID()+".pid")
}

func (l *Link) LogFile() string {
	return filepath.Join(AccessDir, l.ID()+".log")
}

func (l *Link) Start() error {
	pid := l.FindPid()
	l.out.Info("Link.Start: older pid:%d", pid)
	if pid > 0 {
		if ok := libol.HasProcess(pid); ok {
			l.out.Info("OpenVPN.Start: already running")
			return nil
		}
	}
	libol.Go(func() {
		file := l.ConfFile()
		args := []string{
			"-conf", file,
		}
		l.out.Info("%s with %s", l.Path(), args)
		cmd := exec.Command(l.Path(), args...)
		if err := cmd.Start(); err != nil {
			l.out.Error("Link.Start %s: %s", l.uuid, err)
		}
		cmd.Wait()
	})
	return nil
}

func (l *Link) FindPid() int {
	pid := 0
	if v, err := os.ReadFile(l.PidFile()); err == nil {
		fmt.Sscanf(string(v), "%d", &pid)
	}
	return pid
}

func (l *Link) Clean() {
	files := []string{
		l.LogFile(), l.StatusFile(), l.PidFile(), l.ConfFile(),
	}
	for _, file := range files {
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				l.out.Warn("Link.Clean %s", err)
			}
		}
	}
}

func (l *Link) Stop() error {
	if pid := l.FindPid(); pid > 0 {
		l.out.Info("Link.Stop: without stoping: %d", pid)
	} else {
		l.Clean()
	}
	return nil
}

type Links struct {
	lock  sync.RWMutex
	links map[string]*Link
}

func NewLinks() *Links {
	return &Links{
		links: make(map[string]*Link),
	}
}

func (ls *Links) Add(l *Link) {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	ls.links[l.cfg.Connection] = l
}

func (ls *Links) Remove(addr string) *Link {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	if p, ok := ls.links[addr]; ok {
		p.Stop()
		delete(ls.links, addr)
		return p
	}
	return nil
}

type LinuxLink struct {
	link nl.Link
}

func (ll *LinuxLink) Start() error {
	if err := nl.LinkAdd(ll.link); err != nil {
		return err
	}
	return nil
}

func (ll *LinuxLink) Stop() error {
	if err := nl.LinkDel(ll.link); err != nil {
		return err
	}
	return nil
}
