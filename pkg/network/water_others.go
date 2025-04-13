//go:build !linux && !darwin && !windows

package network

import "github.com/luscis/openlan/pkg/water"

func WaterNew(c TapConfig) (*water.Interface, error) {
	return nil, nil
}
