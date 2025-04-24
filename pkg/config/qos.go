package config

import "github.com/luscis/openlan/pkg/libol"

type Qos struct {
	File   string               `json:"-" yaml:"-"`
	Name   string               `json:"name" yaml:"name"`
	Config map[string]*QosLimit `json:"qos,omitempty" yaml:"qos,omitempty"`
}

func (q *Qos) Correct(sw *Switch) {
	for _, rule := range q.Config {
		rule.Correct()
	}
	if q.File == "" {
		q.File = sw.Dir("qos", q.Name)
	}
}

func (q *Qos) Save() {
	if err := libol.MarshalSave(q, q.File, true); err != nil {
		libol.Error("Switch.Save.Qos %s %s", q.Name, err)
	}
}

type QosLimit struct {
	InSpeed float64 `json:"inSpeed,omitempty" yaml:"inSpeed,omitempty"`
}

func (ql *QosLimit) Correct() {
}
