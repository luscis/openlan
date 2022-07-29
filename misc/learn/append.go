package main

import "fmt"

func main() {
	var a = []int{1, 2, 3}

	fmt.Println(a)
	a0 := append(a, []int{4, 5, 6}...)
	a0[0] = 9
	a1 := append(a, []int{7, 8}...)
	fmt.Println(a, a0)
	fmt.Println(a, a1)

	a0 = append(a[:3], []int{4, 5, 6}...)
	a0[0] = 9
	a1 = append(a[:3], []int{7, 8}...)
	fmt.Println(a, a0)
	fmt.Println(a, a1)

	a = make([]int, 0, 1024)

	b := append(a, []int{4, 5, 6}...)
	fmt.Println(b, a)
	//fmt.Println(cap(b), len(b))
	//fmt.Println(cap(a), len(a))

	c := append(b, []int{8, 9}...)
	c[1] = 10
	b[0] = 9
	fmt.Println(c, b)
	//fmt.Println(cap(c), len(c))
	//fmt.Println(cap(b), len(b))

	bb := b
	b = append(b, []int{8, 9}...)
	bb[2] = 11
	fmt.Println(b, bb)
	//fmt.Println(cap(bb), len(bb))
	//fmt.Println(cap(b), len(b))

}
