package libol

import (
	"testing"
)

func TestStartUDP_C(t *testing.T) {
	StartUDP(84209, 4500, "180.109.49.146")
}
