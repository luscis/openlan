package cache

import (
	"github.com/luscis/openlan/pkg/config"
)

func Init(cfg *config.Limit) {
	Access.Init(cfg.Access)
	User.Init(cfg.User)
}
