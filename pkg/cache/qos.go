package cache

import (
	"github.com/luscis/openlan/pkg/libol"
)

type qos struct {
	QosConfig *libol.SafeStrMap
}

var pos = &qos{
	QosConfig: libol.NewSafeStrMap(1024),
}
