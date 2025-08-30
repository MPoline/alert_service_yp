package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/agent/services"
	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/logging"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/MPoline/alert_service_yp/pkg/buildinfo"
	"go.uber.org/zap"
)

var (
	neсMetrics = []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
		"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
		"HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse",
		"MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackSys", "Sys", "TotalAlloc",
	}
	memStorage = storage.NewMemStorage()
	wg         sync.WaitGroup
	sendCh     = make(chan []models.Metrics, 50)
)

func main() {
	buildinfo.Print("Agent")
	fmt.Println("Agent started")

	logger, err := logging.InitLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	logger.Info("Run agent")

	flags.ParseFlags()

	if flags.FlagCryptoKey != "" {
		logger.Info("Initializing encryption", zap.String("public_key", flags.FlagCryptoKey))

		publicKey, err := crypto.LoadPublicKey(flags.FlagCryptoKey)
		if err != nil {
			logger.Error("Failed to load public key",
				zap.String("path", flags.FlagCryptoKey),
				zap.Error(err))
			os.Exit(1)
		}

		if err := services.InitEncryption(publicKey); err != nil {
			logger.Error("Failed to initialize encryption", zap.Error(err))
			os.Exit(1)
		}

		logger.Info("Encryption initialized successfully")
	} else {
		logger.Info("Encryption disabled - no crypto key provided")
	}

	pollInterval := time.Duration(flags.FlagPollInterval) * time.Second
	reportInterval := time.Duration(flags.FlagReportInterval) * time.Second

	wg.Add(3)

	go func() {
		defer wg.Done()
		var workersWG sync.WaitGroup
		workersWG.Add(int(flags.FlagRateLimit))

		for i := 0; i < int(flags.FlagRateLimit); i++ {
			go func(id int) {
				defer workersWG.Done()
				for metrics := range sendCh {
					services.SendMetrics(memStorage, metrics)
				}
			}(i)
		}
		workersWG.Wait()
	}()

	// Сбор метрик
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for range ticker.C {
			services.GetMetrics(memStorage, neсMetrics)
		}
	}()

	// Отправка метрик
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for range ticker.C {
			metricStorage := services.CreateMetrics(memStorage)
			sendCh <- metricStorage
		}
	}()
	wg.Wait()
	close(sendCh)
}
