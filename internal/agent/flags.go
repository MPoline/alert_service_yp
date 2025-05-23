package flags

import (
	"flag"
	"os"
	"strconv"

	"go.uber.org/zap"
)

var (
	FlagRunAddr        string
	FlagReportInterval int64
	FlagPollInterval   int64
)

func ParseFlags() {
	var err error
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&FlagReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.Int64Var(&FlagPollInterval, "p", 2, "frequency of polling metrics")
	flag.Parse()

	if flag.NArg() > 0 {
		zap.L().Info("Error: unknown flag(s)")
		flag.Usage()
		return
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		zap.L().Info("ADDRESS: ", zap.String("envRunAddr", envRunAddr))
		FlagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		FlagReportInterval, err = strconv.ParseInt(envReportInterval, 10, 64)
		if err != nil {
			zap.L().Info("Error parse REPORT_INTERVAL", zap.Error(err))
		}
	}
	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		FlagPollInterval, err = strconv.ParseInt(envPoolInterval, 10, 64)
		if err != nil {
			zap.L().Info("Error parse POLL_INTERVAL", zap.Error(err))
		}
	}
}
