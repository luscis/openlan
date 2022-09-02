package libol

import (
	"crypto/tls"
	"fmt"
	"github.com/go-ldap/ldap"
)

type LDAPConfig struct {
	Server    string
	BindDN    string
	Password  string
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
	if err = conn.Bind(cfg.BindDN, cfg.Password); err != nil {
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
	if len(result.Entries) != 1 {
		return false, fmt.Errorf("invalid users")
	}
	obj := result.Entries[0]
	Debug("LDAPService.Login %v", obj)
	if err = l.Conn.Bind(obj.DN, password); err != nil {
		return false, err
	}
	return true, nil
}
