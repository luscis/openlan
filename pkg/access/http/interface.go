package http

import "github.com/luscis/openlan/pkg/config"

type Accesser interface {
	UUID() string
	Config() *config.Access
}
