package main

func main() {
	ch := make(chan int, 2)

	ch <- 1
	ch <- 2
	ch <- 3

}
