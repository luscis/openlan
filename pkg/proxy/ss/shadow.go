package ss

import (
	"encoding/base64"
	"log"

	c "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/shadowsocks/go-shadowsocks2/core"
)

type ShadowSocks struct {
	Server     string
	Key        string
	Cipher     string
	Password   string
	Plugin     string
	PluginOpts string
	Protocol   string
	out        *libol.SubLogger
}

func NewShadowSocks(cfg *c.ShadowProxy) *ShadowSocks {
	proto := cfg.Protocol
	if proto == "" {
		proto = "tcp"
	}
	return &ShadowSocks{
		Server:     cfg.Server,
		Key:        cfg.Key,
		Cipher:     cfg.Cipher,
		Password:   cfg.Secret,
		Plugin:     cfg.Plugin,
		PluginOpts: cfg.PluginOpts,
		Protocol:   proto,
		out:        libol.NewSubLogger(cfg.Server),
	}
}

func (s *ShadowSocks) Start() {
	var key []byte
	if s.Key != "" {
		k, err := base64.URLEncoding.DecodeString(s.Key)
		if err != nil {
			log.Fatal(err)
		}
		key = k
	}
	addr := s.Server
	cipher := s.Cipher
	password := s.Password
	var err error

	udpAddr := addr
	if s.Plugin != "" {
		addr, err = startPlugin(s.Plugin, s.PluginOpts, addr, true)
		if err != nil {
			log.Fatal(err)
		}
	}
	ciph, err := core.PickCipher(cipher, key, password)
	if err != nil {
		log.Fatal(err)
	}

	if s.Protocol == "udp" {
		go udpRemote(udpAddr, ciph.PacketConn)
	} else {
		go tcpRemote(addr, ciph.StreamConn)
	}
	s.out.Info("ShadowSocks.start %s:%s", s.Protocol, s.Server)
}

func (s *ShadowSocks) Stop() {
	killPlugin()
}
