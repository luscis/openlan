package config

type FindHop struct {
	Name      string      `json:"-" yaml:"-"`
	Check     string      `json:"check" yaml:"check"`
	Params    PingParams  `json:"params" yaml:"params"`
	Mode      string      `json:"mode,omitempty" yaml:"mode,omitempty"`
	NextHop   []string    `json:"nexthop" yaml:"nexthop"`
	Available []MultiPath `json:"available,omitempty" yaml:"available,omitempty"`
	Vrf       string      `json:"-" yaml:"-"`
}

func (ng *FindHop) Correct() {
	if ng.Available == nil {
		ng.Available = []MultiPath{}
	}
	if ng.Check == "" {
		ng.Check = "ping"
	}
	if ng.Mode == "" {
		ng.Mode = "active-backup"
	}
	ng.Params.Correct()
}

type PingParams struct {
	Count    int `json:"count" yaml:"count"`
	Loss     int `json:"loss,omitempty" yaml:"loss,omitempty"`
	Rtt      int `json:"rtt,omitempty" yaml:"rtt,omitempty"`
	Interval int `json:"interval,omitempty" yaml:"interval,omitempty"`
}

func (pp *PingParams) Correct() {
	if pp.Count == 0 {
		pp.Count = 3
	}
	if pp.Loss == 0 {
		pp.Loss = 2
	}
}
