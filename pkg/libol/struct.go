package libol

import "fmt"

type Prefix struct {
	Link     string
	Dst      string
	Src      string
	Gw       string
	Protocol string
	Priority int
	Table    int
}

type Neighbor struct {
	Link    string
	Address string
	HwAddr  string
	State   string
}

type ConnStats struct {
	Total int
	TCP   int
	UDP   int
	ICMP  int
}

func (c ConnStats) String() string {
	return fmt.Sprintf("total:%d|tcp:%d|udp:%d|icmp:%d", c.Total, c.TCP, c.UDP, c.ICMP)
}
