package network

import (
	"github.com/luscis/openlan/pkg/water"
)

func WaterNew(c TapConfig) (dev *water.Interface, err error) {
	deviceType := water.DeviceType(water.TAP)
	if c.Type == TUN {
		deviceType = water.TUN
	}
	cfg := water.Config{DeviceType: deviceType}
	if c.Name == "" {
		return water.New(cfg)
	}
	cfg.PlatformSpecificParams = water.PlatformSpecificParams{
		ComponentID:   "root\\tap0901",
		InterfaceName: c.Name,
		Network:       c.Network,
	}
	if dev, err = water.New(cfg); err == nil {
		return dev, nil
	}
	// try again.
	cfg.PlatformSpecificParams = water.PlatformSpecificParams{
		ComponentID:   "tap0901",
		InterfaceName: c.Name,
		Network:       c.Network,
	}
	if dev, err = water.New(cfg); err == nil {
		return dev, nil
	}
	return nil, err
}
