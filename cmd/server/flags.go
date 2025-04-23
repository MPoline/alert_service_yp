package main

import (
	"flag"
	"fmt"
)

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Error: unknown flag(s)")
		flag.Usage()
		return
	}
}
