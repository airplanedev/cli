package main

import (
	"flag"
	"fmt"
)

var (
	name = flag.String("name", "World", "Name to echo.")
)

func main() {
	flag.Parse()
	fmt.Println("Hello,", *name+"!")
}
