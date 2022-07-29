package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	dip := net.ParseIP("192.168.7.30")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 8888}
	dstAddr := &net.UDPAddr{IP: dip, Port: 9999}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	data := make([]byte, 4096)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i)
	}

	for {
		fmt.Printf("% x ... % x\n", data[:16], data[4080:4096])
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
}
