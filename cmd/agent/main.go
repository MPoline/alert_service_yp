package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
)

func main() {
	buildinfo.Print("Agent")
	fmt.Println("Agent started")

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

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

	sendCh := make(chan []models.Metrics, 50)
	var wg sync.WaitGroup

	sendCtx, cancelSend := context.WithCancel(context.Background())
	defer cancelSend()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(sendCh)

		var workersWG sync.WaitGroup
		workersWG.Add(int(flags.FlagRateLimit))

		for i := 0; i < int(flags.FlagRateLimit); i++ {
			go func(id int) {
				defer workersWG.Done()
				for {
					select {
					case metrics, ok := <-sendCh:
						if !ok {
							logger.Debug("Worker stopped - channel closed", zap.Int("worker_id", id))
							return
						}
						if metrics != nil {
							services.SendMetrics(memStorage, metrics)
						}
					case <-sendCtx.Done():
						logger.Debug("Worker stopped - context cancelled", zap.Int("worker_id", id))
						return
					}
				}
			}(i)
		}
		workersWG.Wait()
		logger.Info("All workers stopped")
	}()

	// Сбор метрик
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				services.GetMetrics(memStorage, neсMetrics)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Отправка метрик
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metricStorage := services.CreateMetrics(memStorage)
				select {
				case sendCh <- metricStorage:
				case <-sendCtx.Done():
					return
				case <-ctx.Done():
					return
				default:
					logger.Warn("Channel full, skipping metrics batch")
				}

			case <-ctx.Done():
				metricStorage := services.CreateMetrics(memStorage)

				select {
				case sendCh <- metricStorage:
					logger.Info("Last metrics sent successfully")
				case <-time.After(100 * time.Millisecond):
					logger.Warn("Failed to send last metrics - timeout")
				case <-sendCtx.Done():
					logger.Warn("Failed to send last metrics - send context cancelled")
				}

				cancelSend()
				return
			}
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down agent gracefully...")

	wg.Wait()

	logger.Info("Agent stopped")
}
