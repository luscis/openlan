package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"sync"
	"sync/atomic"
	"time"
)

func test(limit rate.Limit, burst int, size uint32, wg *sync.WaitGroup) {
	var (
		numOK   = uint32(0)
		numFail = uint32(0)
	)

	// Very slow replenishing bucket.
	lim := rate.NewLimiter(limit, burst)

	now := time.Now().Unix()
	at := time.Now().Add(15 * time.Second)
	// Tries to take a token, atomically updates the counter and decreases the wait
	// group counter.

	f := func() {
		if ok := lim.AllowN(time.Now(), int(size)); ok {
			//fmt.Printf("%d\n", time.Now().Unix())
			atomic.AddUint32(&numOK, size)
		} else {
			atomic.AddUint32(&numFail, size)
		}
	}

	for at.After(time.Now()) {
		go f()
	}
	dt := time.Now().Unix() - now
	fmt.Printf("size = %d rate: %d\n", size, numOK/uint32(dt))
	wg.Done()
}

func main() {
	const (
		limit       = 10 * 1024
		burst       = 10 * 1024 * 2
		numRequests = uint32(50)
	)

	wg := &sync.WaitGroup{}
	wg.Add(int(numRequests))
	for i := uint32(0); i < numRequests; i++ {
		go test(limit, burst, 64+(i*64), wg)
	}
	wg.Wait()
}
