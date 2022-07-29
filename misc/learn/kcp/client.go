package main

import (
	"github.com/xtaci/kcp-go/v5"
	"time"
)

func main() {
	conn, err := kcp.DialWithOptions("192.168.7.30:9999", nil, 10, 3)
	if err != nil {
		panic(err)
	}

	for {
		data := make([]byte, 4096)
		_, _ = conn.Write(data)
		time.Sleep(time.Second)
	}
}
