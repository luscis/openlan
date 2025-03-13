package proxy

import (
	"time"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/socks5"
)

type SocksProxy struct {
	server *socks5.Server
	out    *libol.SubLogger
	cfg    *config.SocksProxy
}

func NewSocksProxy(cfg *config.SocksProxy) *SocksProxy {
	s := &SocksProxy{
		cfg: cfg,
		out: libol.NewSubLogger(cfg.Listen),
	}
	// Create a SOCKS5 server
	auth := cfg.Auth
	authMethods := make([]socks5.Authenticator, 0, 2)
	if auth != nil && len(auth.Username) > 0 {
		author := socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				auth.Username: auth.Password,
			},
		}
		authMethods = append(authMethods, author)

	}
	conf := &socks5.Config{
		Backends:    cfg.Backends,
		AuthMethods: authMethods,
		Logger:      s.out,
	}
	server, err := socks5.New(conf)
	if err != nil {
		s.out.Error("NewSocksProxy %s", err)
		return nil
	}
	s.server = server
	return s
}

func (s *SocksProxy) Start() {
	if s.server == nil || s.cfg == nil {
		return
	}
	addr := s.cfg.Listen
	s.out.Info("Proxy.startSocks")

	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		if err := s.server.ListenAndServe("tcp", addr); err != nil {
			s.out.Warn("Proxy.startSocks %s", err)
			return err
		}
		return nil
	})
}
