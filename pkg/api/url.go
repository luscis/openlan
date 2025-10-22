package api

import "github.com/gorilla/mux"

func Add(router *mux.Router, cs SwitchApi) {
	Link{cs: cs}.Router(router)
	User{}.Router(router)
	Bgp{}.Router(router)
	IPSec{}.Router(router)
	Prefix{}.Router(router)
	Neighbor{}.Router(router)
	Access{}.Router(router)
	OnLine{}.Router(router)
	Lease{}.Router(router)
	Server{cs: cs}.Router(router)
	Device{}.Router(router)
	VPNClient{}.Router(router)
	PProf{}.Router(router)
	Config{cs: cs}.Router(router)
	Version{}.Router(router)
	Log{}.Router(router)
	OpenAPI{}.Router(router)
	ZTrust{}.Router(router)
	Qos{}.Router(router)
	Output{cs: cs}.Router(router)
	ACL{}.Router(router)
	Route{cs: cs}.Router(router)
	FindHop{}.Router(router)
	Rate{cs: cs}.Router(router)
	SNAT{}.Router(router)
	DNAT{}.Router(router)
	Network{cs: cs}.Router(router)
}
