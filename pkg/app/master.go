package app

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/network"
)

type Master interface {
	UUID() string
	Protocol() string
	OffClient(client libol.SocketClient)
	ReadTap(device network.Taper, readAt func(f *libol.FrameMessage) error)
	NewTap(tenant string) (network.Taper, error)
}
