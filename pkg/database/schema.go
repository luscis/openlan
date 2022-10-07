package database

import (
	"strconv"
)

type Switch struct {
	UUID            string            `ovsdb:"_uuid" json:"uuid"`
	Protocol        string            `ovsdb:"protocol" json:"protocol"`
	Listen          int               `ovsdb:"listen" json:"listen"`
	OtherConfig     map[string]string `ovsdb:"other_config" json:"other_config"`
	VirtualNetworks []string          `ovsdb:"virtual_networks" json:"virtual_networks"`
}

type VirtualNetwork struct {
	UUID         string            `ovsdb:"_uuid" json:"uuid"`
	Name         string            `ovsdb:"name" json:"name"`
	Provider     string            `ovsdb:"provider" json:"provider"`
	Bridge       string            `ovsdb:"bridge" json:"bridge"`
	Address      string            `ovsdb:"address" json:"address"`
	OtherConfig  map[string]string `ovsdb:"other_config" json:"other_config"`
	RemoteLinks  []string          `ovsdb:"remote_links" json:"remote_links"`
	LocalLinks   []string          `ovsdb:"local_links" json:"local_links"`
	OpenVPN      *string           `ovsdb:"open_vpn" json:"open_vpn"`
	PrefixRoutes []string          `ovsdb:"prefix_routes" json:"prefix_routes"`
}

type VirtualLink struct {
	UUID           string            `ovsdb:"_uuid" json:"uuid"`
	Network        string            `ovsdb:"network" json:"network"`
	Connection     string            `ovsdb:"connection" json:"connection"`
	Device         string            `ovsdb:"device" json:"device"`
	OtherConfig    map[string]string `ovsdb:"other_config" json:"other_config"`
	Authentication map[string]string `ovsdb:"authentication" json:"authentication"`
	LinkState      string            `ovsdb:"link_state" json:"link_state"`
	Status         map[string]string `ovsdb:"status" json:"status"`
}

func (l *VirtualLink) IsUdpIn() bool {
	if HasPrefix(l.Device, 4, "spi:") &&
		HasPrefix(l.Connection, 4, "udp:") {
		return true
	}
	return false
}

func (l *VirtualLink) Spi() uint32 {
	spi, _ := strconv.Atoi(l.Device[4:])
	return uint32(spi)
}

type OpenVPN struct {
	UUID     string `ovsdb:"_uuid" json:"uuid"`
	Protocol string `ovsdb:"protocol" json:"protocol"`
	Listen   int    `ovsdb:"listen" json:"listen"`
}

type NameCache struct {
	UUID     string `ovsdb:"_uuid" json:"uuid"`
	Name     string `ovsdb:"name" json:"name"`
	Address  string `ovsdb:"address" json:"address"`
	UpdateAt string `ovsdb:"update_at" json:"update_at"`
}

type PrefixRoute struct {
	UUID    string `ovsdb:"_uuid" json:"uuid"`
	Network string `ovsdb:"network" json:"network"`
	Prefix  string `ovsdb:"prefix" json:"prefix"`
	Source  string `ovsdb:"source" json:"source"`
	Gateway string `ovsdb:"gateway" json:"gateway"`
	Mode    string `ovsdb:"mode" json:"mode"`
}
