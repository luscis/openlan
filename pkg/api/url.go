package api

import "github.com/gorilla/mux"

func Add(router *mux.Router, switcher Switcher) {
	Link{Switcher: switcher}.Router(router)
	User{}.Router(router)
	Neighbor{}.Router(router)
	Point{}.Router(router)
	Network{}.Router(router)
	OnLine{}.Router(router)
	Lease{}.Router(router)
	Server{Switcher: switcher}.Router(router)
	Device{}.Router(router)
	VPNClient{}.Router(router)
	PProf{}.Router(router)
	VxLAN{}.Router(router)
	Esp{}.Router(router)
	EspState{}.Router(router)
	EspPolicy{}.Router(router)
	Config{Switcher: switcher}.Router(router)
	Version{}.Router(router)
}
