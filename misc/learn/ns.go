package main

import (
	"fmt"
	"github.com/vishvananda/netns"
)

func main() {
	ns, err := netns.GetFromName("hi")
	fmt.Println(ns, err)
	ns, err = netns.GetFromName("dan")
	fmt.Println(ns, err)
}
