package libol

import (
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap"
)

type LDAPConfig struct {
	Server    string
	BindUser  string
	BindPass  string
	BaseDN    string
	Attr      string
	Filter    string
	EnableTls bool
	Timeout   int64
}

type LDAPService struct {
	Conn  *ldap.Conn
	Cfg   LDAPConfig
	Error string
}

func NewLDAPService(cfg LDAPConfig) (*LDAPService, error) {
	conn, err := ldap.Dial("tcp", cfg.Server)
	if err != nil {
		return nil, err
	}
	if cfg.EnableTls {
		err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return nil, err
		}
	}
	if err = conn.Bind(cfg.BindUser, cfg.BindPass); err != nil {
		return nil, err
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 8 * 3600
	}
	return &LDAPService{Conn: conn, Cfg: cfg}, nil
}

func (l *LDAPService) Login(userName, password string) (bool, error) {
	cfg := l.Cfg
	if err := l.Conn.Bind(cfg.BindUser, cfg.BindPass); err != nil {
		Error("LDAPService.Login bind %v: %s", err)
		return false, nil
	}

	request := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false,
		"("+fmt.Sprintf(cfg.Filter, userName)+")",
		[]string{cfg.Attr}, nil,
	)
	result, err := l.Conn.Search(request)
	if err != nil {
		return false, err
	}
	if len(result.Entries) <= 0 {
		return false, fmt.Errorf("User not found")
	}

	obj := result.Entries[0]
	if err = l.Conn.Bind(obj.DN, password); err != nil {
		return false, err
	}
	return true, nil
}

func (l *LDAPService) State() string {
	if l.Conn.IsClosing() {
		return "closing"
	}
	return "success"
}
