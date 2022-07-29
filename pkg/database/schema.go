package database

type Switch struct {
	UUID            string            `ovsdb:"_uuid"`
	Protocol        string            `ovsdb:"protocol"`
	Listen          int               `ovsdb:"listen"`
	OtherConfig     map[string]string `ovsdb:"other_config" yaml:"other_config"`
	VirtualNetworks []string          `ovsdb:"virtual_networks" yaml:"virtual_networks"`
}

type VirtualNetwork struct {
	UUID         string            `ovsdb:"_uuid"`
	Name         string            `ovsdb:"name"`
	Provider     string            `ovsdb:"provider"`
	Bridge       string            `ovsdb:"bridge"`
	Address      string            `ovsdb:"address"`
	OtherConfig  map[string]string `ovsdb:"other_config" yaml:"other_config"`
	RemoteLinks  []string          `ovsdb:"remote_links" yaml:"remote_links"`
	LocalLinks   []string          `ovsdb:"local_links" yaml:"local_links"`
	OpenVPN      *string           `ovsdb:"open_vpn" yaml:"open_vpn"`
	PrefixRoutes []string          `ovsdb:"prefix_routes" yaml:"prefix_routes"`
}

type VirtualLink struct {
	UUID           string            `ovsdb:"_uuid"`
	Network        string            `ovsdb:"network"`
	Connection     string            `ovsdb:"connection"`
	Device         string            `ovsdb:"device"`
	OtherConfig    map[string]string `ovsdb:"other_config" yaml:"other_config"`
	Authentication map[string]string `ovsdb:"authentication" yaml:"authentication"`
	LinkState      string            `ovsdb:"link_state" yaml:"link_state"`
	Status         map[string]string `ovsdb:"status" yaml:"status"`
}

type OpenVPN struct {
	UUID     string `ovsdb:"_uuid"`
	Protocol string `ovsdb:"protocol"`
	Listen   int    `ovsdb:"listen"`
}

type NameCache struct {
	UUID     string `ovsdb:"_uuid"`
	Name     string `ovsdb:"name"`
	Address  string `ovsdb:"address"`
	UpdateAt string `ovsdb:"update_at" yaml:"update_at"`
}

type PrefixRoute struct {
	UUID    string `ovsdb:"_uuid"`
	Network string `ovsdb:"network"`
	Prefix  string `ovsdb:"prefix"`
	Source  string `ovsdb:"source"`
	Gateway string `ovsdb:"gateway"`
	Mode    string `ovsdb:"mode"`
}
