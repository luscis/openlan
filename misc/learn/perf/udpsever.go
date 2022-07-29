package main

import (
	"fmt"
	"github.com/songgao/water"
	"net"
)

func main() {
	var remote *net.UDPAddr

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 9981})
	if err != nil {
		fmt.Println(err)
		return
	}
	device, err := water.New(water.Config{DeviceType: water.TAP})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Local: <%s> \n", device.Name())
	fmt.Printf("Local: <%s> \n", listener.LocalAddr().String())

	//1500-20-8-8, 16 = 1448
	data := make([]byte, 1448+16+8)
	go func() {
		for {
			n, remoteAddr, err := listener.ReadFromUDP(data)
			if err != nil {
				fmt.Printf("error during read: %s", err)
			}

			if n == 0 {
				continue
			}
			fmt.Printf("<%s> %d\n", remoteAddr, n)
			remote = remoteAddr
			//fmt.Printf("<%s> %s\n", remoteAddr, data[:n])
			_, err = device.Write(data[8:n])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	//udpMtu := 1500-20-8 //1472
	frameData := make([]byte, 1448+16+8)
	//header := make([]byte, 8)
	for {
		n, err := device.Read(frameData[8:])
		if err != nil {
			break
		}

		fmt.Printf("<%s> %d %x\n", device.Name(), n, frameData[:20])
		if n == 0 || remote == nil {
			continue
		}

		_, err = listener.WriteToUDP(frameData[:n+8], remote)
		if err != nil {
			fmt.Println(err)
		}
	}
}
