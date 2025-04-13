//go:build !linux

package network

type VRF struct {
	name  string
	table int
}

func NewVRF(name string, table int) *VRF {
	return &VRF{}
}

func (v *VRF) Table() int {
	return v.table
}

func (v *VRF) Name() string {
	return v.name
}
