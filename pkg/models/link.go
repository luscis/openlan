package models

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

type Link struct {
	User       string
	Network    string
	Protocol   string
	StatusFile string
}

func (l *Link) reload() *schema.Access {
	status := &schema.Access{}
	_ = libol.UnmarshalLoad(status, l.StatusFile)
	return status
}

func (l *Link) Status() *schema.Access {
	return l.reload()
}
