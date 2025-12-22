package cache

import (
	"github.com/luscis/openlan/pkg/config"
)

func Init(cfg *config.Limit) {
	Access.Init(cfg.Access)
	Neighbor.Init(cfg.Neighbor)
	Online.Init(cfg.OnLine)
	User.Init(cfg.User)
}

func Reload() {
}
