package http

import "github.com/luscis/openlan/pkg/config"

type Pointer interface {
	UUID() string
	Config() *config.Point
}
