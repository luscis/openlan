package main

import (
	"fmt"
	"github.com/songgao/water"
	"net"
)

func main() {
	sip := net.ParseIP("192.168.4.151")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: sip, Port: 9981}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}

	device, err := water.New(water.Config{DeviceType: water.TAP})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Local: <%s> \n", device.Name())

	frameData := make([]byte, 1448+16+8) //1472
	go func() {
		for {
			n, err := device.Read(frameData[8:])
			if err != nil {
				break
			}
			if n == 0 || conn == nil {
				continue
			}

			fmt.Printf("<%s> %d\n", device.Name(), n)
			fmt.Printf("<%s> % x\n", device.Name(), frameData[:20])

			_, err = conn.Write(frameData[:n+8])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	data := make([]byte, 1448+16+8)
	for {
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		if n == 0 {
			continue
		}
		fmt.Printf("<%s> %x\n", dstAddr.String(), data[:n])
		_, err = device.Write(data[8:n])
		if err != nil {
			fmt.Println(err)
		}
	}

	conn.Close()
	device.Close()
}
