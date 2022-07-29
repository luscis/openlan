package main

import (
	"fmt"
	"github.com/luscis/openlan/pkg/libol"
	"os"
)

func main() {
	cfg := libol.LDAPConfig{}
	cfg.Server = os.Getenv("LDAPServer")
	cfg.Password = os.Getenv("LDAPPassword")
	cfg.BaseDN = os.Getenv("LDAPBaseDN")
	cfg.BindDN = os.Getenv("LDAPBindDN")
	cfg.Filter = os.Getenv("LDAPFilter")
	cfg.Attr = os.Getenv("LDAPAttr")

	if l, err := libol.NewLDAPService(cfg); err != nil {
		panic(err)
	} else {
		username := os.Getenv("username")
		password := os.Getenv("password")
		if ok, err := l.Login(username, password); !ok {
			panic(err)
		} else {
			fmt.Println("success")
		}
	}
	fmt.Println(cfg)
}
