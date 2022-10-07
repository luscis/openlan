package libol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOpenUDP_C(t *testing.T) {
	udp := &UdpInServer{Port: 4500}
	err := udp.Open()
	assert.Equal(t, nil, err, "has not error")
	assert.NotEqual(t, -1, udp.Socket, "valid socket")

	go func() {
		conn, err := udp.Recv()
		fmt.Println(conn, err)
	}()

	err = udp.Send(&UdpInConnection{
		Spi:        84209,
		RemoteAddr: "180.109.49.146",
		RemotePort: 4500,
	})
	assert.Equal(t, nil, err, "has not error")
	time.Sleep(time.Second * 2)
}
