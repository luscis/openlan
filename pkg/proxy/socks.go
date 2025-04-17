package proxy

import (
	"crypto/tls"
	"time"

	"github.com/luscis/openlan/pkg/config"
	co "github.com/luscis/openlan/pkg/config"
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
	s.Initialize()
	return s
}

func (s *SocksProxy) Initialize() {
	user, pass := co.SplitSecret(s.cfg.Secret)
	authMethods := make([]socks5.Authenticator, 0, 2)
	if user != "" {
		author := socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				user: pass,
			},
		}
		authMethods = append(authMethods, author)
		s.out.Debug("SocksProxy: Auth user %s", user)
	}
	conf := &socks5.Config{
		Backends:    s.cfg.Backends,
		AuthMethods: authMethods,
		Logger:      s.out,
	}
	crt := s.cfg.Cert
	if crt != nil && crt.KeyFile != "" {
		conf.TlsConfig = &tls.Config{
			Certificates: crt.GetCertificates(),
		}
	}
	server, err := socks5.New(conf)
	if err != nil {
		s.out.Error("NewSocksProxy %s", err)
	}
	s.server = server
}

func (s *SocksProxy) Start() {
	if s.server == nil || s.cfg == nil {
		return
	}
	addr := s.cfg.Listen

	crt := s.cfg.Cert
	if crt == nil || crt.KeyFile == "" {
		s.out.Info("SocksProxy.Start: socks5://%s", s.cfg.Listen)
	} else {
		s.out.Info("SocksProxy.Start: sockss://%s", s.cfg.Listen)
	}

	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		if err := s.server.ListenAndServe("tcp", addr); err != nil {
			s.out.Warn("SocksProxy.Start %s", err)
			return err
		}
		return nil
	})
}

func (s *SocksProxy) Stop() {
	if s.server != nil {
		s.server = nil
	}
}

func (s *SocksProxy) Save() {
}
