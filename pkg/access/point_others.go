//go:build !linux && !windows && !darwin

package access

import "github.com/luscis/openlan/pkg/config"

type Point struct {
}

func NewPoint(config *config.Point) *Point {
	return nil
}

func (p *Point) Initialize() {

}

func (p *Point) Start() {

}
