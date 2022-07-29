package _switch

import (
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	OlapBin = "openlan-point"
	OlapDir = "/var/openlan/point"
)

type Link struct {
	cfg    *co.Point
	status *schema.Point
	out    *libol.SubLogger
	uuid   string
}

func NewLink(uuid string, cfg *co.Point) *Link {
	return &Link{
		uuid: uuid,
		cfg:  cfg,
		out:  libol.NewSubLogger(cfg.Network),
	}
}

func (l *Link) Model() *models.Link {
	cfg := l.Conf()
	return &models.Link{
		User:       cfg.Username,
		Network:    cfg.Network,
		Protocol:   cfg.Protocol,
		StatusFile: l.StatusFile(),
	}
}

func (l *Link) Initialize() {
	file := l.ConfFile()
	l.cfg.StatusFile = l.StatusFile()
	l.cfg.PidFile = l.PidFile()
	_ = libol.MarshalSave(l.cfg, file, true)
}

func (l *Link) Conf() *co.Point {
	return l.cfg
}

func (l *Link) UUID() string {
	return l.uuid
}

func (l *Link) Path() string {
	return OlapBin
}

func (l *Link) ConfFile() string {
	return filepath.Join(OlapDir, l.uuid+".json")
}

func (l *Link) StatusFile() string {
	return filepath.Join(OlapDir, l.uuid+".status")
}

func (l *Link) PidFile() string {
	return filepath.Join(OlapDir, l.uuid+".pid")
}

func (l *Link) LogFile() string {
	return filepath.Join(OlapDir, l.uuid+".log")
}

func (l *Link) Start() {
	file := l.ConfFile()
	log, err := libol.CreateFile(l.LogFile())
	if err != nil {
		l.out.Warn("Link.Start %s", err)
		return
	}
	libol.Go(func() {
		args := []string{
			"-alias", l.cfg.Network,
			"-conn", l.cfg.Connection,
			"-conf", file,
			"-terminal", "ww",
		}
		l.out.Debug("Link.Start %s %v", l.Path(), args)
		cmd := exec.Command(l.Path(), args...)
		cmd.Stdout = log
		cmd.Stderr = log
		if err := cmd.Run(); err != nil {
			l.out.Error("Link.Start %s: %s", l.uuid, err)
		}
	})
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

func (l *Link) Stop() {
	if data, err := ioutil.ReadFile(l.PidFile()); err != nil {
		l.out.Debug("Link.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			l.out.Warn("Link.Stop %s: %s", pid, err)
		}
	}
	l.Clean()
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
