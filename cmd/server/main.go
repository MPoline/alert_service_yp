package main

import (
	"fmt"
	"os"
	"time"

	"github.com/MPoline/alert_service_yp/internal/logging"
	"github.com/MPoline/alert_service_yp/internal/server/api"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"go.uber.org/zap"
)

func main() {
	logger, err := logging.InitLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	r := api.InitRouter()

	flags.ParseFlags()

	var storageType string

	if flags.FlagDatabaseDSN == "" {
		storageType = "memory"
	} else {
		storageType = "database"
	}

	storage.InitStorage(storageType)
	logger.Info("Storage type: ", zap.String("storageType", storageType))

	if storageType == "memory" {
		if flags.FlagRestore {
			err := storage.LoadFromFile(storage.MetricStorage, flags.FlagFileStoragePath)
			if err != nil {
				logger.Warn("Error read from file: ", zap.Error(err))
			}
		}

		storeInterval := time.Second * time.Duration(flags.FlagStoreInterval)
		if flags.FlagStoreInterval > 0 {
			ticker := time.NewTicker(storeInterval)
			go func() {
				for range ticker.C {
					storage.SaveToFile(storage.MetricStorage, flags.FlagFileStoragePath)
				}
			}()
		}

	}

	defer storage.MetricStorage.Close()

	err = r.Run(flags.FlagRunAddr)
	if err != nil {
		logger.Warn("Error start server: ", zap.Error(err))
	}
}
