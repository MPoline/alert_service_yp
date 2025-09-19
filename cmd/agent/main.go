package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/agent/services"
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

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		zap.L().Warn("Failed to get local IP address, using fallback", zap.Error(err))
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

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

	localIP := getLocalIP()
	logger.Info("Using local IP",
		zap.String("ip", localIP))

	clientManager, err := services.NewClientManager()
	if err != nil {
		logger.Error("Failed to initialize client manager", zap.Error(err))
		os.Exit(1)
	}
	defer clientManager.Close()

	logger.Info("Client manager initialized successfully")

	pollInterval := time.Duration(flags.FlagPollInterval) * time.Second
	reportInterval := time.Duration(flags.FlagReportInterval) * time.Second

	sendCh := make(chan []models.Metrics, flags.FlagRateLimit)
	var wg sync.WaitGroup

	sendCtx, cancelSend := context.WithCancel(ctx)
	defer cancelSend()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var workersWG sync.WaitGroup
		workersWG.Add(int(flags.FlagRateLimit))

		for i := 0; i < int(flags.FlagRateLimit); i++ {
			go func(id int) {
				defer workersWG.Done()
				for metrics := range sendCh {
					if metrics != nil {
						clientManager.SendMetrics(memStorage, metrics, localIP)
					}
				}
				logger.Debug("Worker stopped - channel closed",
					zap.Int("worker_id", id),
					zap.String("worker_ip", localIP))
			}(i)
		}
		workersWG.Wait()
		logger.Info("All workers stopped", zap.String("agent_ip", localIP))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := clientManager.HealthCheck(); err != nil {
					logger.Warn("Health check failed",
						zap.Error(err),
						zap.String("agent_ip", localIP))
				} else {
					logger.Debug("Health check passed",
						zap.String("agent_ip", localIP))
				}

			case <-ctx.Done():
				logger.Info("Health check stopped",
					zap.String("agent_ip", localIP))
				return
			}
		}
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
				logger.Info("Metrics collection stopped", zap.String("agent_ip", localIP))
				return
			}
		}
	}()

	// Отправка метрик
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(sendCh)

		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metricStorage := services.CreateMetrics(memStorage)
				select {
				case sendCh <- metricStorage:
					logger.Debug("Metrics batch sent to channel",
						zap.Int("metrics_count", len(metricStorage)),
						zap.String("agent_ip", localIP))
				case <-sendCtx.Done():
					logger.Info("Send context cancelled", zap.String("agent_ip", localIP))
					return
				case <-ctx.Done():
					logger.Info("Main context cancelled", zap.String("agent_ip", localIP))
					return
				default:
					logger.Warn("Channel full, skipping metrics batch",
						zap.String("agent_ip", localIP))
				}

			case <-ctx.Done():
				logger.Info("Shutdown initiated, sending final metrics",
					zap.String("agent_ip", localIP))

				metricStorage := services.CreateMetrics(memStorage)

				select {
				case sendCh <- metricStorage:
					logger.Info("Last metrics sent successfully",
						zap.Int("metrics_count", len(metricStorage)),
						zap.String("agent_ip", localIP))
				case <-time.After(100 * time.Millisecond):
					logger.Warn("Failed to send last metrics - timeout",
						zap.String("agent_ip", localIP))
				case <-sendCtx.Done():
					logger.Warn("Failed to send last metrics - send context cancelled",
						zap.String("agent_ip", localIP))
				}

				cancelSend()
				return
			}
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down agent gracefully...",
		zap.String("agent_ip", localIP))

	wg.Wait()

	logger.Info("Agent stopped", zap.String("agent_ip", localIP))
}
