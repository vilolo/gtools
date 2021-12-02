package main

import (
	"flag"
	"fmt"
)

func main() {
	fmt.Println("m2 start !!!")

	id := flag.Int("id", 0, "id")
	fmt.Println(*id)
}
