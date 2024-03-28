package config

import "github.com/luscis/openlan/pkg/libol"

type Qos struct {
	File   string               `json:"file"`
	Name   string               `json:"name"`
	Config map[string]*QosLimit `json:"qos,omitempty"`
}

func (q *Qos) Save() {
	if err := libol.MarshalSave(q, q.File, true); err != nil {
		libol.Error("Switch.Save.Qos %s %s", q.Name, err)
	}
}

type QosLimit struct {
	InSpeed int64 `json:"inSpeed,omitempty"`
}

func (ql *QosLimit) Correct() {
}
