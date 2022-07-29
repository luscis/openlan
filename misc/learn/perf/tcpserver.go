package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/songgao/water"
	"net"
)

// 40MB on 1000Mb

func ReadFull(conn net.Conn, buffer []byte) error {
	offset := 0
	left := len(buffer)

	for left > 0 {
		tmp := make([]byte, left)
		n, err := conn.Read(tmp)
		if err != nil {
			return err
		}
		copy(buffer[offset:], tmp)
		offset += n
		left -= n
	}
	return nil
}

func WriteFull(conn net.Conn, buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	for left > 0 {
		tmp := buffer[offset:]
		n, err := conn.Write(tmp)
		if err != nil {
			return err
		}
		offset += n
		left = size - offset
	}
	return nil
}

func Client(addr string) {
	srcAddr := &net.TCPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.DialTCP("tcp", srcAddr, dstAddr)
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

	go func() {
		frameData := make([]byte, 1600+4)

		for {
			n, err := device.Read(frameData[4:])
			if err != nil {
				break
			}
			if n == 0 || conn == nil {
				continue
			}

			binary.BigEndian.PutUint16(frameData[2:4], uint16(n))
			//fmt.Printf("<%s> %d\n", device.Name(), n)
			//fmt.Printf("<%s> % x\n", device.Name(), frameData[:20])
			err = WriteFull(conn, frameData[:n+4])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	for {
		data := make([]byte, 1600+4)

		err := ReadFull(conn, data[:4])
		if err != nil {
			fmt.Printf("error during read: %s", err)
			break
		}

		size := binary.BigEndian.Uint16(data[2:4])
		if size == 0 || size > 1600 {
			continue
		}

		err = ReadFull(conn, data[4:size+4])
		if err != nil {
			fmt.Printf("error during read: %s", err)
			break
		}

		_, err = device.Write(data[4 : size+4])
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	_ = conn.Close()
	_ = device.Close()
}

func Server(addr string) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
	}
	device, err := water.New(water.Config{DeviceType: water.TAP})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Local : <%s> \n", device.Name())
	fmt.Printf("Remote: <%s> \n", conn.LocalAddr().String())

	go func() {
		data := make([]byte, 1600+4) //MTU:1500, 1500+14+4

		for {
			err := ReadFull(conn, data[:4])
			if err != nil {
				fmt.Printf("error during read: %s", err)
			}

			size := binary.BigEndian.Uint16(data[2:4])
			if size == 0 || size > 1600 {
				continue
			}

			//fmt.Printf("%d %x\n", size, data[:20])
			err = ReadFull(conn, data[4:size+4])
			if err != nil {
				fmt.Printf("error during read: %s", err)
			}

			_, err = device.Write(data[4 : size+4])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	for {
		frameData := make([]byte, 1600+4)

		n, err := device.Read(frameData[4:])
		if err != nil {
			break
		}

		binary.BigEndian.PutUint16(frameData[2:4], uint16(n))
		if n == 0 {
			continue
		}

		//fmt.Printf("<%s> %d %x\n", device.Name(), n, frameData[:20])
		err = WriteFull(conn, frameData[:n+4])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	address := "127.0.0.1:9981"
	mode := "server"
	flag.StringVar(&address, "addr", address, "the address listen.")
	flag.StringVar(&mode, "mode", mode, "client or server.")
	flag.Parse()

	if mode == "server" {
		go Server(address)
	} else if mode == "client" {
		go Client(address)
	}
	libol.Wait()
}
