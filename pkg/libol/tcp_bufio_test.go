package libol

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
)

type Counter struct {
	Rx int
	Tx int
}

func handleConnection(conn net.Conn, n int, c *Counter) {
	for i := 0; i < n; i++ {
		// will listen for message to process ending in newline (\n)
		message, _ := bufio.NewReader(conn).ReadString('\n')
		if len(message) == 0 {
			break
		}
		c.Rx += 1
		// output message received
		//fmt.Printf("Server Received: %s", string(message))
		// sample process for string received
		newMessage := strings.ToUpper(message)
		// send new string back to client
		conn.Write([]byte(newMessage))
	}
}

func startServer(wg *sync.WaitGroup, ok chan int, n int, c *Counter) {
	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	ok <- 1

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}

	handleConnection(conn, n, c)
	conn.Close()

	wg.Done()
}

func startClient(wg *sync.WaitGroup, ok chan int, n int, c *Counter) {
	<-ok

	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8081")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		for {
			message, _ := bufio.NewReader(conn).ReadString('\n')
			//fmt.Printf("Client Received: %s", string(message))
			if message != "" {
				break
			}
		}
		conn.Close()
		wg.Done()
	}()

	go func() {
		for i := 0; i < n; i++ {
			fmt.Fprintf(conn, "From the client\n")
			c.Tx += 1
		}
	}()
}

func TestClientAndServer(t *testing.T) {
	wg := &sync.WaitGroup{}
	ok := make(chan int, 1)

	c := &Counter{}
	wg.Add(1)
	go startServer(wg, ok, 128, c)
	wg.Add(1)
	go startClient(wg, ok, 128, c)

	wg.Wait()
	//fmt.Printf("Total tx: %d, rx: %d\n", c.Tx, c.Rx)
}

func BenchmarkClientAndServer(b *testing.B) {
	wg := &sync.WaitGroup{}
	ok := make(chan int, 1)

	c := &Counter{}
	wg.Add(1)
	go startServer(wg, ok, b.N, c)
	wg.Add(1)
	go startClient(wg, ok, b.N, c)

	wg.Wait()
	//fmt.Printf("Total tx: %d, rx: %d\n", c.Tx, c.Rx)
}
