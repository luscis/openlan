package main

import (
	"fmt"
)

func main() {
	c := make(map[int]map[int]string, 32)

	for i := 0; i < 64; i++ {
		c[i] = make(map[int]string, 2)
	}

	fmt.Printf("%d,%s\n", len(c), c)
}
