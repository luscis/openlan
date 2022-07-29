package main

import (
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"io"
	"net"
)

func main() {
	fmt.Println("kcp listens on 10000")
	lis, err := kcp.ListenWithOptions(":10000", nil, 10, 3)
	if err != nil {
		panic(err)
	}
	for {
		conn, e := lis.AcceptKCP()
		if e != nil {
			panic(e)
		}
		go func(conn net.Conn) {
			var buffer = make([]byte, 4096)
			for {
				n, e := conn.Read(buffer)
				if e != nil {
					if e == io.EOF {
						fmt.Println("receive EOF")
						break
					}
					fmt.Println(e)
					break
				}
				fmt.Println("receive from client:", buffer[:n])
			}
		}(conn)
	}
}
