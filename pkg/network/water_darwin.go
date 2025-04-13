package network

import (
	"github.com/luscis/openlan/pkg/water"
)

func WaterNew(c TapConfig) (*water.Interface, error) {
	deviceType := water.DeviceType(water.TAP)
	if c.Type == TUN {
		deviceType = water.TUN
	}
	cfg := water.Config{DeviceType: deviceType}
	if c.Name != "" {
		cfg.PlatformSpecificParams = water.PlatformSpecificParams{
			Name: c.Name,
		}
	}
	return water.New(cfg)
}
