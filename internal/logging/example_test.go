package logging_test

import (
	"testing"

	"github.com/MPoline/alert_service_yp/internal/logging"
	"go.uber.org/zap"
)

func ExampleInitLog() {
	logger, err := logging.InitLog()
	if err != nil {
		panic(err)
	}
	defer logging.Sync()

	logger.Info("Logger initialized successfully")
}

func TestLogging(t *testing.T) {
	t.Run("Initialize logger", func(t *testing.T) {
		logger, err := logging.InitLog()
		if err != nil {
			t.Fatalf("InitLog failed: %v", err)
		}
		defer logging.Sync()

		if logger == nil {
			t.Error("Logger should not be nil")
		}
	})

	t.Run("Check logger level", func(t *testing.T) {
		logger, _ := logging.InitLog()
		defer logging.Sync()

		if ce := logger.Check(zap.InfoLevel, "test"); ce == nil {
			t.Error("Logger should support InfoLevel")
		}
	})
}

func ExampleSync() {
	logger, err := logging.InitLog()
	if err != nil {
		panic(err)
	}

	defer logging.Sync()

	logger.Info("This log message will be flushed by Sync")
}
