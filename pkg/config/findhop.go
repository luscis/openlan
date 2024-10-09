package config

type FindHop struct {
	Name      string      `json:"-"`
	Check     string      `json:"check"`
	Params    PingParams  `json:"params"`
	Mode      string      `json:"mode,omitempty"`
	NextHop   []string    `json:"nexthop"`
	Available []MultiPath `json:"available,omitempty"`
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
	Count    int `json:"count"`
	Loss     int `json:"loss,omitempty"`
	Rtt      int `json:"rtt,omitempty"`
	Interval int `json:"interval,omitempty"`
}

func (pp *PingParams) Correct() {
	if pp.Count == 0 {
		pp.Count = 3
	}
	if pp.Loss == 0 {
		pp.Loss = 2
	}
}
