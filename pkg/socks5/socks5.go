package socks5

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"golang.org/x/net/context"
)

const (
	socks5Version = uint8(5)
)

// Config is used to setup and configure a Server
type Config struct {
	// AuthMethods can be provided to implement custom authentication
	// By default, "auth-less" mode is enabled.
	// For password-based auth use UserPassAuthenticator.
	AuthMethods []Authenticator

	// If provided, username/password authentication is enabled,
	// by appending a UserPassAuthenticator to AuthMethods. If not provided,
	// and AUthMethods is nil, then "auth-less" mode is enabled.
	Credentials CredentialStore

	// Resolver can be provided to do custom name resolution.
	// Defaults to DNSResolver if not provided.
	Resolver NameResolver

	// Rules is provided to enable custom logic around permitting
	// various commands. If not provided, PermitAll is used.
	Rules RuleSet

	// Rewriter can be used to transparently rewrite addresses.
	// This is invoked before the RuleSet is invoked.
	// Defaults to NoRewrite.
	Rewriter AddressRewriter

	// BindIP is used for bind or udp associate
	BindIP net.IP

	// Logger can be used to provide a custom log target.
	// Defaults to stdout.
	Logger *libol.SubLogger

	// Optional function for dialing out
	Dial func(ctx context.Context, network, addr string) (net.Conn, error)

	// Backends forwarding socks request
	Backends co.FindBackend

	// TLS Configurations
	TlsConfig *tls.Config
}

// Server is reponsible for accepting connections and handling
// the details of the SOCKS5 protocol
type Server struct {
	config      *Config
	authMethods map[uint8]Authenticator
}

// New creates a new Server and potentially returns an error
func New(conf *Config) (*Server, error) {
	// Ensure we have at least one authentication method enabled
	if len(conf.AuthMethods) == 0 {
		if conf.Credentials != nil {
			conf.AuthMethods = []Authenticator{&UserPassAuthenticator{conf.Credentials}}
		} else {
			conf.AuthMethods = []Authenticator{&NoAuthAuthenticator{}}
		}
	}

	// Ensure we have a DNS resolver
	if conf.Resolver == nil {
		conf.Resolver = DNSResolver{}
	}

	// Ensure we have a rule set
	if conf.Rules == nil {
		conf.Rules = PermitAll()
	}

	// Ensure we have a log target
	if conf.Logger == nil {
		conf.Logger = libol.NewSubLogger("")
	}

	server := &Server{
		config: conf,
	}

	server.authMethods = make(map[uint8]Authenticator)

	for _, a := range conf.AuthMethods {
		server.authMethods[a.GetCode()] = a
	}

	return server, nil
}

// ListenAndServe is used to create a listener and serve on it
func (s *Server) ListenAndServe(network, addr string) error {
	var l net.Listener
	var err error

	if s.config.TlsConfig != nil {
		l, err = tls.Listen(network, addr, s.config.TlsConfig)
	} else {
		l, err = net.Listen(network, addr)
	}
	if err != nil {
		return err
	}

	return s.Serve(l)
}

// Serve is used to serve connections from a listener
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.ServeConn(conn)
	}
}

// ServeConn is used to serve a single connection.
func (s *Server) ServeConn(conn net.Conn) error {
	defer conn.Close()
	bufConn := bufio.NewReader(conn)

	// Read the version byte
	version := []byte{0}
	if _, err := bufConn.Read(version); err != nil {
		s.config.Logger.Error("Socks.ServeConn Failed to get version byte: %v", err)
		return err
	}

	// Ensure we are compatible
	if version[0] != socks5Version {
		err := fmt.Errorf("Unsupported SOCKS version: %v", version)
		s.config.Logger.Error("Socks.ServeConn %v", err)
		return err
	}

	// Authenticate the connection
	authContext, err := s.authenticate(conn, bufConn)
	if err != nil {
		err = fmt.Errorf("Failed to authenticate: %v", err)
		s.config.Logger.Error("Socks.ServeConn %v", err)
		return err
	}

	request, err := NewRequest(bufConn)
	if err != nil {
		if err == unrecognizedAddrType {
			if err := sendReply(conn, addrTypeNotSupported, nil); err != nil {
				return fmt.Errorf("Failed to send reply: %v", err)
			}
		}
		return fmt.Errorf("Failed to read destination address: %v", err)
	}
	request.AuthContext = authContext
	if client, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		request.RemoteAddr = &AddrSpec{IP: client.IP, Port: client.Port}
	}

	dstAddr := request.DestAddr
	if s.config.Backends != nil {
		via := s.config.Backends.FindBackend(dstAddr.Address())
		if via != nil {
			if err := s.toForward(request, conn, via); err != nil {
				s.config.Logger.Error("Socks.ServeConn: %v", err)
				return err
			}
			return nil
		}
	}

	s.config.Logger.Info("Socks.ServeConn CONNECT %s", dstAddr.Address())
	//Process the client request
	if err := s.handleRequest(request, conn); err != nil {
		err = fmt.Errorf("Failed to handle request: %v", err)
		s.config.Logger.Error("Socks.ServeConn %v", err)
	}

	return nil
}

func (s *Server) toTunnel(local net.Conn, target net.Conn) {
	defer local.Close()
	defer target.Close()
	wait := libol.NewWaitOne(2)

	libol.Go(func() {
		defer wait.Done()
		io.Copy(local, target)
	})
	libol.Go(func() {
		defer wait.Done()
		io.Copy(target, local)
	})
	wait.Wait()
}

func (s *Server) openConn(via *co.HttpForward) (net.Conn, error) {
	remote := via.SocksAddr()

	if via.Protocol == "tls" || via.Protocol == "https" {
		conf := &tls.Config{
			InsecureSkipVerify: via.Insecure,
		}
		conf.RootCAs, _ = co.GetCertPool(via.CaCert)
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		return tls.DialWithDialer(dialer, "tcp", remote, conf)
	}
	return net.DialTimeout("tcp", remote, 10*time.Second)
}

func (s *Server) toForward(req *Request, local net.Conn, via *co.HttpForward) error {
	dstAddr := req.DestAddr
	proxy := via.SocksAddr()

	s.config.Logger.Info("Socks.ServeConn CONNECT %s via %s", dstAddr, proxy)

	target, err := s.openConn(via)
	if err != nil {
		sendReply(local, networkUnreachable, nil)
		s.config.Logger.Error("Socks.ServeConn CONNECT %s: unreachable", dstAddr)
		return err
	}

	// Handshake: SOCKS5 auth.
	if via.Secret != "" {
		_, err = target.Write([]byte{socks5Version, 1, UserPassAuth})
	} else {
		_, err = target.Write([]byte{socks5Version, 1, NoAuth})
	}
	if err != nil {
		sendReply(local, serverFailure, nil)
		return err
	}

	// Auth request: User password.
	if via.Secret != "" {
		header := []byte{0, 0}
		_, err := target.Read(header)
		if header[0] != socks5Version || header[1] != UserPassAuth {
			s.config.Logger.Error("Socks.ServeConn CONNECT %s: wrong %s", dstAddr, header[1])
			sendReply(local, serverFailure, nil)
			return err
		}

		user, pass := co.SplitSecret(via.Secret)
		auth := []byte{userAuthVersion}
		auth = append(auth, byte(len(user)))
		auth = append(auth, user...)
		auth = append(auth, byte(len(pass)))
		auth = append(auth, pass...)
		if _, err = target.Write(auth); err != nil {
			sendReply(local, serverFailure, nil)
			return err
		}
		// Check UserAuth reply.
		header = []byte{0, 0}
		_, err = target.Read(header)
		if header[0] != userAuthVersion || header[1] != successReply {
			sendReply(local, serverFailure, nil)
			s.config.Logger.Error("Socks.ServeConn CONNECT %s: auth %d", dstAddr, header[1])
			return err
		}
	} else {
		// Check Socks5 reply.
		header := []byte{0, 0}
		_, err = target.Read(header)
		if header[0] != socks5Version || header[1] != successReply {
			sendReply(local, serverFailure, nil)
			s.config.Logger.Error("Socks.ServeConn CONNECT %s: socks5 %d", dstAddr, header[1])
			return err
		}
	}

	// Bind: Connect Domain and port number.
	domain := []byte(dstAddr.FQDN)
	port := []byte{0, 0}
	binary.BigEndian.PutUint16(port, uint16(dstAddr.Port))
	bind := []byte{socks5Version, 1, 0, 3}
	bind = append(bind, byte(len(domain)))
	bind = append(bind, domain...)
	bind = append(bind, port...)
	if _, err := target.Write(bind); err != nil {
		sendReply(local, serverFailure, nil)
		return err
	}

	s.toTunnel(local, target)
	return nil
}

func (s *Server) SetBackends(find co.FindBackend) {
	s.config.Backends = find
}
