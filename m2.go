package main

import (
	"flag"
	"fmt"
	"./strategys"
)

func main() {
	fmt.Println("m2 start !!!")

	t := flag.Int("t", 0, "type")
	flag.Parse()
	// fmt.Println(*t)

	//检测
	if *t == 0 {
		strategys.M()
	}
}
