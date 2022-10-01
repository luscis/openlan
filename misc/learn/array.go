package main

import (
	"encoding/json"
	"fmt"
)

func InArray(data []int) {
	data[0] = 0x04
	fmt.Println(data)
}

func main() {
	var a = []int{1, 2, 3}

	fmt.Println(a)
	InArray(a)
	fmt.Println(a)

	b := a
	a[1] = 5
	fmt.Println(a, b)

	b[1] = 6
	fmt.Println(a, b)

	c := a[1:]
	a[2] = 10
	fmt.Println(a, c)

	c[1] = 11
	fmt.Println(a, c)

	var aa []int
	str := `[1, 2, 3]`
	err := json.Unmarshal([]byte(str), &aa)
	fmt.Println(err)
	fmt.Println(aa)
}
