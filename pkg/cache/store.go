package cache

import (
	"github.com/luscis/openlan/pkg/config"
)

func Init(cfg *config.Perf) {
	Access.Init(cfg.Access)
	Link.Init(cfg.Link)
	Neighbor.Init(cfg.Neighbor)
	Online.Init(cfg.OnLine)
	User.Init(cfg.User)
}

func Reload() {
}
