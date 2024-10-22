package api

import "github.com/gorilla/mux"

func Add(router *mux.Router, switcher Switcher) {
	Link{Switcher: switcher}.Router(router)
	User{}.Router(router)
	Neighbor{}.Router(router)
	Point{}.Router(router)
	Network{Switcher: switcher}.Router(router)
	OnLine{}.Router(router)
	Lease{}.Router(router)
	Server{Switcher: switcher}.Router(router)
	Device{}.Router(router)
	VPNClient{}.Router(router)
	PProf{}.Router(router)
	Config{Switcher: switcher}.Router(router)
	Version{}.Router(router)
	Log{}.Router(router)
	OpenAPI{}.Router(router)
	ZTrust{}.Router(router)
	QosApi{}.Router(router)
	Output{Switcher: switcher}.Router(router)
	ACL{}.Router(router)
	Route{Switcher: switcher}.Router(router)
	IPSec{}.Router(router)
	FindHop{}.Router(router)
	Rate{Switcher: switcher}.Router(router)
	SNAT{}.Router(router)
}
