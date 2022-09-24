package cache

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

type user struct {
	Lock    sync.RWMutex
	File    string
	Cert    string
	Users   *libol.SafeStrMap
	LdapCfg *libol.LDAPConfig
	LdapSvc *libol.LDAPService
}

func (w *user) Load() {
	file := w.File
	reader, err := libol.OpenRead(file)
	if err != nil {
		return
	}
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.SplitN(line, ":", 4)
		if len(columns) < 2 {
			continue
		}
		user := columns[0]
		pass := columns[1]
		role := "guest"
		leStr := ""
		if len(columns) > 2 {
			role = columns[2]
		}
		if len(columns) > 3 {
			leStr = columns[3]
		}
		lease, _ := libol.GetLocalTime(libol.LeaseTime, leStr)
		obj := &models.User{
			Name:     user,
			Password: pass,
			Role:     role,
			Lease:    lease,
		}
		obj.Update()
		w.Add(obj)
	}
	if err := scanner.Err(); err != nil {
		libol.Warn("User.Load %v", err)
	}
}

func (w *user) Save() error {
	if w.File == "" {
		return nil
	}
	fp, err := libol.OpenTrunk(w.File)
	if err != nil {
		return err
	}
	for obj := range w.List() {
		if obj == nil {
			break
		}
		if obj.Role == "ldap" {
			continue
		}
		line := obj.Id()
		line += ":" + obj.Password
		line += ":" + obj.Role
		line += ":" + obj.Lease.Format(libol.LeaseTime)
		_, _ = fp.WriteString(line + "\n")
	}
	return nil
}

func (w *user) SetFile(value string) {
	w.File = value
}

func (w *user) Init(size int) {
	w.Users = libol.NewSafeStrMap(size)
}

func (w *user) Add(user *models.User) {
	libol.Debug("user.Add %v", user)
	key := user.Id()
	if older := w.Get(key); older == nil {
		_ = w.Users.Set(key, user)
	} else { // Update pass and role.
		if user.Role != "" {
			older.Role = user.Role
		}
		if user.Password != "" {
			older.Password = user.Password
		}
		if user.Alias != "" {
			older.Alias = user.Alias
		}
		older.UpdateAt = user.UpdateAt
		if !user.Lease.IsZero() {
			older.Lease = user.Lease
		}
	}
}

func (w *user) Del(key string) {
	libol.Debug("user.Add %s", key)
	w.Users.Del(key)
}

func (w *user) Get(key string) *models.User {
	if v := w.Users.Get(key); v != nil {
		return v.(*models.User)
	}
	return nil
}

func (w *user) List() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.Users.Iter(func(k string, v interface{}) {
			c <- v.(*models.User)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *user) CheckLdap(obj *models.User) *models.User {
	svc := w.GetLdap()
	if svc == nil {
		return nil
	}
	u := w.Get(obj.Id())
	libol.Debug("CheckLdap %s", u)
	if u != nil && u.Role != "ldap" {
		return nil
	}
	if ok, err := svc.Login(obj.Id(), obj.Password); !ok {
		libol.Warn("CheckLdap %s", err)
		return nil
	}
	user := &models.User{
		Name:     obj.Id(),
		Password: obj.Password,
		Role:     "ldap",
		Alias:    obj.Alias,
	}
	user.Update()
	w.Add(user)
	return user
}

func (w *user) Timeout(user *models.User) bool {
	if user.Role == "ldap" {
		return time.Now().Unix()-user.UpdateAt > w.LdapCfg.Timeout
	}
	return true
}

func (w *user) Check(obj *models.User) (*models.User, error) {
	if w.Cert != "" {
		pemData, err := ioutil.ReadFile(w.Cert)
		if err != nil {
			return nil, err
		}
		block, rest := pem.Decode(pemData)
		if block == nil || len(rest) > 0 {
			return nil, libol.NewErr("certificate decoding error")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		if now.Before(cert.NotBefore) {
			return nil, libol.NewErr("certificate isn't yet valid")
		} else if now.After(cert.NotAfter) {
			return nil, libol.NewErr("certificate has expired")
		}
	}
	if u := w.Get(obj.Id()); u != nil {
		if u.Role == "" || u.Role == "admin" || u.Role == "guest" {
			if u.Password == obj.Password {
				t0 := time.Now()
				t1 := u.Lease
				if t1.Year() < 2000 || t1.After(t0) {
					return u, nil
				}
				return nil, libol.NewErr("out of date")
			}
		}
	}
	if u := w.CheckLdap(obj); u != nil {
		return u, nil
	}
	return nil, libol.NewErr("wrong password")
}

func (w *user) GetLdap() *libol.LDAPService {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	if w.LdapCfg == nil {
		return nil
	}
	if w.LdapSvc == nil || w.LdapSvc.Conn.IsClosing() {
		if l, err := libol.NewLDAPService(*w.LdapCfg); err != nil {
			libol.Warn("user.GetLdap %s", err)
			w.LdapSvc = nil
		} else {
			w.LdapSvc = l
		}
	}
	return w.LdapSvc
}

func (w *user) SetLdap(cfg *libol.LDAPConfig) {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	if w.LdapCfg != cfg {
		w.LdapCfg = cfg
	}
	if l, err := libol.NewLDAPService(*cfg); err != nil {
		libol.Warn("user.SetLdap %s", err)
	} else {
		libol.Info("user.SetLdap %s", w.LdapCfg.Server)
		w.LdapSvc = l
	}
}

func (w *user) SetCert(cfg *libol.CertConfig) {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	w.Cert = cfg.Crt
}

var User = user{
	Users: libol.NewSafeStrMap(1024),
}
