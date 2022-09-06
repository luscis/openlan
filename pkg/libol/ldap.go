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
	Conn *ldap.Conn
	Cfg  LDAPConfig
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
	request := ldap.NewSearchRequest(
		l.Cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf(l.Cfg.Filter, userName),
		[]string{l.Cfg.Attr},
		nil,
	)
	Debug("LDAPService.Login %v", request)
	result, err := l.Conn.Search(request)
	if err != nil {
		return false, err
	}
	if len(result.Entries) <= 0 {
		return false, fmt.Errorf("user not found")
	}
	obj := result.Entries[0]
	Debug("LDAPService.Login %v", obj)
	if err = l.Conn.Bind(obj.DN, password); err != nil {
		return false, err
	}
	if err = l.Conn.Bind(l.Cfg.BindUser, l.Cfg.BindPass); err != nil {
		return false, nil
	}
	return true, nil
}
