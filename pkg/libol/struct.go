package libol

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
