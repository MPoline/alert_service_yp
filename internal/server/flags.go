package server

import (
	"flag"
	"os"
	"strconv"

	"go.uber.org/zap"
)

var (
	FlagRunAddr         string
	FlagStoreInterval   int64
	FlagFileStoragePath string
	FlagRestore         bool
)

func ParseFlags() {
	var err error
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&FlagStoreInterval, "i", 300, "frequency of save metrics")
	flag.StringVar(&FlagFileStoragePath, "f", "./savedMetrics", "address of file for save metrics")
	flag.BoolVar(&FlagRestore, "r", false, "read metrics from file")
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

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		FlagStoreInterval, err = strconv.ParseInt(envStoreInterval, 10, 64)
		if err != nil {
			zap.L().Info("Error parse STORE_INTERVAL", zap.Error(err))
		}
	}

	if envStorePath := os.Getenv("FILE_STORAGE_PATH"); envStorePath != "" {
		zap.L().Info("FILE_STORAGE_PATH: ", zap.String("envStorePath", envStorePath))
		FlagFileStoragePath = envStorePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		FlagRestore, err = strconv.ParseBool(envRestore)
		if err != nil {
			zap.L().Info("Error parse RESTORE", zap.Error(err))
		}
	}
}
