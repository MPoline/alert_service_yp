package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	FlagRunAddr         string
	FlagStoreInterval   int64
	FlagFileStoragePath string
	FlagRestore         bool
)

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&FlagStoreInterval, "i", 300, "frequency of save metrics")
	flag.StringVar(&FlagFileStoragePath, "f", "./savedMetrics", "address of file for save metrics")
	flag.BoolVar(&FlagRestore, "r", false, "read metrics from file")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Error: unknown flag(s)")
		flag.Usage()
		return
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		FlagStoreInterval, _ = strconv.ParseInt(envStoreInterval, 10, 64)
	}

	if envStorePath := os.Getenv("FILE_STORAGE_PATH"); envStorePath != "" {
		FlagFileStoragePath = envStorePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		FlagRestore, _ = strconv.ParseBool(envRestore)
	}
}
