package cache

import (
	"fmt"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	"github.com/stretchr/testify/assert"
	"testing"
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
		UUID:     "fake",
		Client: &SocketClientMock{},
	})
	assert.Equal(t, 1 , Point.Clients.Len(), "MUST be same")
	EspState.Add(&models.EspState{
		EspState: &schema.EspState {},
	})
	assert.Equal(t, 1 , EspState.State.Len(), "MUST be same")
	User.Add(&models.User{
		Alias:    "fake",
	})
	assert.Equal(t, 1 , User.Users.Len(), "MUST be same")
	Link.Add("fake-uuid", &models.Link{
		User:       "fake",
	})
	assert.Equal(t, 1 , Link.Links.Len(), "MUST be same")
	EspPolicy.Add(&models.EspPolicy{
		EspPolicy: &schema.EspPolicy{},
	})
	assert.Equal(t, 1 , EspPolicy.Policy.Len(), "MUST be same")
	Online.Add(&models.Line{
		EthType:    0,
	})
	assert.Equal(t, 1 , Online.Lines.Len(), "MUST be same")
	Neighbor.Add(&models.Neighbor{
		Network: "fake",
	})
	assert.Equal(t, 1 , Neighbor.Neighbors.Len(), "MUST be same")
	Reload()
	assert.Equal(t, 0, EspState.State.Len(), "MUST be same")
	assert.Equal(t, 0 , EspPolicy.Policy.Len(), "MUST be same")
}