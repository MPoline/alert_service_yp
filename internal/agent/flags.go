package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	FlagRunAddr        string
	FlagReportInterval int64
	FlagPollInterval   int64
)

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&FlagReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.Int64Var(&FlagPollInterval, "p", 2, "frequency of polling metrics")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Error: unknown flag(s)")
		flag.Usage()
		return
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		fmt.Println("ADDRESS: ", envRunAddr)
		FlagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		FlagReportInterval, _ = strconv.ParseInt(envReportInterval, 10, 64)
	}
	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		FlagPollInterval, _ = strconv.ParseInt(envPoolInterval, 10, 64)
	}
}
