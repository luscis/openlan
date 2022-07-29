package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 9999})
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		data := make([]byte, 4096*2)
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		fmt.Printf("from %s and %d\n", remoteAddr, n)
		fmt.Printf("% x ... % x\n", data[:16], data[4080:4096])
	}
}
