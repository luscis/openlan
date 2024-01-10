package network

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/vishvananda/netlink"
	nl "github.com/vishvananda/netlink"
)

type VRF struct {
	name  string
	table int
	link  nl.Link
	out   *libol.SubLogger
}

var tableId = 1000

func GenTable() int {
	tableId += 1
	return tableId
}

func NewVRF(name string, table int) *VRF {
	if table == 0 {
		table = GenTable()
	}
	return &VRF{
		name:  name,
		table: table,
		out:   libol.NewSubLogger(name),
	}
}

func (v *VRF) Up() error {
	if link, _ := nl.LinkByName(v.name); link != nil {
		v.link = link
		return nil
	}

	link := &nl.Vrf{
		LinkAttrs: nl.LinkAttrs{
			Name: v.name,
		},
		Table: uint32(v.table),
	}

	if err := nl.LinkAdd(link); err != nil {
		return err
	}
	if err := nl.LinkSetUp(link); err != nil {
		return err
	}

	v.link = link
	v.out.Info("VRF.Up %s", v.name)

	return nil
}

func (v *VRF) Down() error {
	if v.link == nil {
		return nil
	}

	if err := nl.LinkDel(v.link); err != nil {
		return err
	}

	v.link = nil
	v.out.Info("VRF.Down %s", v.name)

	return nil
}

func (v *VRF) AddSlave(name string) error {
	if v.link == nil {
		return libol.NewErr("VRF %s not up", v.name)
	}

	link, err := netlink.LinkByName(name)
	if link == nil {
		return err
	}

	if err := nl.LinkSetMaster(link, v.link); err != nil {
		return nil
	}

	v.out.Info("VRF.AddSlave %s", name)

	return nil
}

func (v *VRF) DelSlave(name string) error {
	link, _ := netlink.LinkByName(name)
	if link == nil {
		return nil
	}

	if err := nl.LinkSetNoMaster(link); err != nil {
		return nil
	}

	v.out.Info("VRF.DelSlave %s", name)

	return nil
}

func (v *VRF) Table() int {
	return v.table
}
