package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	flagRunAddr        string
	flagReportInterval int64
	flagPollInterval   int64
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&flagReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.Int64Var(&flagPollInterval, "p", 2, "frequency of polling metrics")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Error: unknown flag(s)")
		flag.Usage()
		return
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		fmt.Println("ADDRESS: ", envRunAddr)
		flagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		flagReportInterval, _ = strconv.ParseInt(envReportInterval, 10, 64)
	}
	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		flagPollInterval, _ = strconv.ParseInt(envPoolInterval, 10, 64)
	}
}
