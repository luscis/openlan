package config

type NextGroup struct {
	Check            string      `json:"check"`
	Ping             PingParams  `json:"ping"`
	Mode             string      `json:"mode,omitempty"`
	NextHop          []string    `json:"nexthop"`
	AvailableNextHop []MultiPath `json:"availableNexthop,omitempty"`
}

func (ng *NextGroup) Correct() {

	if ng.AvailableNextHop == nil {
		ng.AvailableNextHop = []MultiPath{}
	}
}

type PingParams struct {
	Count          int `json:"count"`
	Loss           int `json:"loss,omitempty"`
	Rtt            int `json:"rtt,omitempty"`
	CheckFrequency int `json:"checkFrequency,omitempty"`
}
