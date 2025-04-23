package main

import (
	"flag"
	"fmt"
)

var (
	flagRunAddr        string
	flagReportInterval int
	flagPollInterval   int
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of polling metrics")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Error: unknown flag(s)")
		flag.Usage()
		return
	}
}
