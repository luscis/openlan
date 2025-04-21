//go:build !linux && !windows && !darwin

package access

import "github.com/luscis/openlan/pkg/config"

type Access struct {
	MixAccess
}

func NewAccess(config *config.Access) *Access {
	return nil
}
