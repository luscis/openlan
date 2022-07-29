package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/songgao/water"
	"net"
)

// 409Mib on 1000Mb

type socket struct {
	conn   net.Conn
	frames int
	buffer []byte
}

func (s *socket) ReadFull() (error, []byte) {
	size := len(s.buffer)
	if size > 0 {
		buf := s.buffer
		if size > 4 {
			ps := binary.BigEndian.Uint16(buf[2:4])
			fs := int(ps) + 4
			//fmt.Printf("fs %d, size %d, % x\n", fs, size, buf)
			if size >= fs {
				s.buffer = buf[fs:]
				return nil, buf[:fs]
			}
		}
	}
	tmp := make([]byte, 1518*s.frames)
	if size > 0 {
		copy(tmp[:size], s.buffer[:size])
	}
	n, err := s.conn.Read(tmp[size:])
	if err != nil {
		return err, nil
	}
	//fmt.Printf("n %d, size %d, % x\n", n, size, s.buffer)
	rs := size + n
	hs := binary.BigEndian.Uint16(tmp[2:4])
	fs := int(hs) + 4
	//fmt.Printf("rs %d, fs %d, % x\n", rs, fs, tmp[:rs])
	if rs >= fs {
		s.buffer = tmp[fs:rs]
		return nil, tmp[:fs]
	} else {
		s.buffer = tmp[:rs]
	}
	return nil, nil
}

func (s *socket) WriteFull(buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	for left > 0 {
		tmp := buffer[offset:]
		n, err := s.conn.Write(tmp)
		if err != nil {
			return err
		}
		offset += n
		left = size - offset
	}
	return nil
}

func xClient(addr string, frames int) {
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

	sock := &socket{
		conn:   conn,
		frames: frames,
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
			err = sock.WriteFull(frameData[:n+4])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	for {
		err, data := sock.ReadFull()
		if err != nil {
			fmt.Printf("error during read: %s", err)
			break
		}
		if data == nil {
			continue
		}
		_, err = device.Write(data[4:])
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	_ = conn.Close()
	_ = device.Close()
}

func xServer(addr string, frames int) {
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

	sock := &socket{
		conn:   conn,
		frames: frames,
	}
	go func() {
		for {
			err, data := sock.ReadFull()
			if err != nil {
				fmt.Printf("error during read: %s", err)
				break
			}
			if data == nil {
				continue
			}
			_, err = device.Write(data[4:])
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
		err = sock.WriteFull(frameData[:n+4])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	address := "127.0.0.1:9981"
	mode := "server"
	frames := 16
	flag.StringVar(&address, "addr", address, "the address listen.")
	flag.StringVar(&mode, "mode", mode, "client or server.")
	flag.IntVar(&frames, "frames", frames, "frames of buffer.")
	flag.Parse()

	if mode == "server" {
		go xServer(address, frames)
	} else if mode == "client" {
		go xClient(address, frames)
	}
	libol.Wait()
}
