package main

import (
	"fmt"
	"vizurth/wildberries-test-l0/quest-24/point"
)

func main() {
	p1 := point.New(10, 2)
	p2 := point.New(12, 15)
	fmt.Println(p1.Distance(p2))
}
