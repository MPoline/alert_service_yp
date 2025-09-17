package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/logging"
	"github.com/MPoline/alert_service_yp/internal/server/api"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/MPoline/alert_service_yp/internal/server/services"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/MPoline/alert_service_yp/pkg/buildinfo"
	"go.uber.org/zap"
)

func main() {
	buildinfo.Print("Server")
	fmt.Println("Server started")

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	logger, err := logging.InitLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	flags.ParseFlags()

	var privateKey *rsa.PrivateKey
	if flags.FlagCryptoKey != "" {
		logger.Info("Initializing decryption", zap.String("private_key", flags.FlagCryptoKey))

		privateKey, err = crypto.LoadPrivateKey(flags.FlagCryptoKey)
		if err != nil {
			logger.Error("Failed to load private key",
				zap.String("path", flags.FlagCryptoKey),
				zap.Error(err))
			os.Exit(1)
		}

		services.InitDecryption(privateKey)
		logger.Info("Decryption initialized successfully")
	} else {
		logger.Info("Decryption disabled - no crypto key provided")
	}

	if flags.FlagGRPCAddress != "" {
		if err := services.InitGRPCServer(privateKey, flags.FlagKey); err != nil {
			logger.Error("Failed to start gRPC server", zap.Error(err))
			os.Exit(1)
		}
		defer services.StopGRPCServer()
	}

	r := api.InitRouter()

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
			defer ticker.Stop()

			go func() {
				for {
					select {
					case <-ticker.C:
						storage.SaveToFile(storage.MetricStorage, flags.FlagFileStoragePath)
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	defer storage.MetricStorage.Close()

	server := &http.Server{
		Addr:    flags.FlagRunAddr,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	logger.Info("Server is running", zap.String("address", flags.FlagRunAddr))

	if flags.FlagGRPCAddress != "" {
		logger.Info("gRPC server is running", zap.String("address", flags.FlagGRPCAddress))
	}

	<-ctx.Done()
	logger.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if storageType == "memory" {
		logger.Info("Saving data to file before shutdown...")
		if err := storage.SaveToFile(storage.MetricStorage, flags.FlagFileStoragePath); err != nil {
			logger.Error("Error saving data to file", zap.Error(err))
		}
	}

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}
