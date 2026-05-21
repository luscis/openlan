package app

import (
	"github.com/luscis/openlan/pkg/libsock"
	"github.com/luscis/openlan/pkg/network"
)

type Master interface {
	UUID() string
	Protocol() string
	OffClient(client libsock.SocketClient)
	ReadTap(device network.Taper, readAt func(f *libsock.FrameMessage) error)
	NewTap(tenant string) (network.Taper, error)
}
