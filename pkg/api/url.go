package api

import "github.com/gorilla/mux"

func Add(router *mux.Router, cs SwitchApi) {
	User{}.Router(router)
	Ldap{cs: cs}.Router(router)
	KernelRoute{}.Router(router)
	KernelNeighbor{}.Router(router)
	Neighbor{}.Router(router)
	Access{}.Router(router)
	OnLine{}.Router(router)
	Lease{}.Router(router)
	Server{cs: cs}.Router(router)
	Device{}.Router(router)
	VPNClient{}.Router(router)
	PProf{}.Router(router)
	Config{cs: cs}.Router(router)
	Version{cs: cs}.Router(router)
	Log{}.Router(router)
	RateLimit{cs: cs}.Router(router)
	Ceci{}.Router(router)
	Bgp{}.Router(router)
	IPSec{}.Router(router)

	OpenAPI{}.Router(router)
	ZTrust{}.Router(router)
	ClientQoS{}.Router(router)
	Output{cs: cs}.Router(router)
	ACL{}.Router(router)
	PrefixRoute{cs: cs}.Router(router)
	FindHop{}.Router(router)
	SNAT{}.Router(router)
	DNAT{}.Router(router)
	RouterTunnel{}.Router(router)
	Network{cs: cs}.Router(router)
}
