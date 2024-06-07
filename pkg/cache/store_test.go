package cache

import (
	"fmt"
	"testing"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/stretchr/testify/assert"
)

type SocketClientMock struct {
	libol.SocketClientImpl
}

func (s *SocketClientMock) String() string {
	return "fake"
}

func TestInit(t *testing.T) {
	cfg := &config.Perf{}
	cfg.Correct()
	Init(cfg)
	fmt.Println(Point)
	Point.Add(&models.Point{
		UUID:   "fake",
		Client: &SocketClientMock{},
	})
	assert.Equal(t, 1, Point.Clients.Len(), "MUST be same")
}
