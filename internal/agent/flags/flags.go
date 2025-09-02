// Package flags предоставляет функциональность для обработки флагов командной строки
// и переменных окружения агента сбора метрик.
//
// Пакет поддерживает:
// - Парсинг флагов командной строки
// - Чтение значений из переменных окружения
// - Приоритет переменных окружения над флагами
// - Валидацию и логирование параметров
package flags

import (
	"flag"
	"os"
	"strconv"

	"github.com/MPoline/alert_service_yp/internal/config"
	"go.uber.org/zap"
)

// Глобальные переменные с параметрами агента
var (
	// FlagRunAddr - адрес и порт сервера (флаг -a, переменная ADDRESS)
	FlagRunAddr string

	// FlagReportInterval - интервал отправки метрик на сервер в секундах (флаг -r, переменная REPORT_INTERVAL)
	FlagReportInterval int64

	// FlagPollInterval - интервал сбора метрик в секундах (флаг -p, переменная POLL_INTERVAL)
	FlagPollInterval int64

	// FlagKey - ключ для подписи данных (флаг -k, переменная KEY)
	FlagKey string

	// FlagRateLimit - лимит одновременных запросов (флаг -l, переменная RATE_LIMIT)
	FlagRateLimit int64

	// FlagCryptoKey - путь до файла с публичным ключом
	FlagCryptoKey string

	FlagConfigFile string
)

// ParseFlags обрабатывает аргументы командной строки и переменные окружения.
// Приоритет значений: переменные окружения > флаги командной строки > значения по умолчанию.
//
// Поддерживаемые флаги:
//
//	-a : адрес сервера (по умолчанию ":8080")
//	-r : интервал отправки метрик в секундах (по умолчанию 10)
//	-p : интервал сбора метрик в секундах (по умолчанию 2)
//	-k : ключ для подписи (по умолчанию "+randomSrting+")
//	-l : лимит одновременных запросов (по умолчанию 5)
//	-с : ассиметричное шифрование (по умолчанию не используется)
//
// Поддерживаемые переменные окружения:
//
//	ADDRESS
//	REPORT_INTERVAL
//	POLL_INTERVAL
//	KEY
//	RATE_LIMIT
//	CRYPTO_KEY
//
// Пример использования:
//
//	flags.ParseFlags()
//	addr := flags.FlagRunAddr
func ParseFlags() {
	var err error
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&FlagReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.Int64Var(&FlagPollInterval, "p", 2, "frequency of polling metrics")
	flag.StringVar(&FlagKey, "k", "+randomSrting+", "key hashSHA256")
	flag.Int64Var(&FlagRateLimit, "l", 5, "rateLimit workers")
	flag.StringVar(&FlagCryptoKey, "crypto-key", "", "path to file with public key for encryption")
	flag.StringVar(&FlagConfigFile, "config", "", "path to configuration file")
	flag.StringVar(&FlagConfigFile, "c", "", "path to configuration file (shorthand)")

	flag.Parse()

	if flag.NArg() > 0 {
		zap.L().Info("Error: unknown flag(s)")
		flag.Usage()
		return
	}

	var fileConfig *config.AgentConfig
	if FlagConfigFile != "" {
		fileConfig, err = config.LoadAgentConfig(FlagConfigFile)
		if err != nil {
			zap.L().Error("Failed to load config file", zap.Error(err))
		}
	}

	if fileConfig != nil {
		applyFileConfig(fileConfig)
	}

	readEnvVars()

	validateAndLogFlags()

	if flag.NArg() > 0 {
		zap.L().Warn("Unknown arguments detected", zap.Strings("args", flag.Args()))
	}
}

func applyFileConfig(config *config.AgentConfig) {
	if FlagRunAddr == "localhost:8080" && config.Address != "" {
		FlagRunAddr = config.Address
	}
	if FlagReportInterval == 10 && config.ReportInterval != 0 {
		FlagReportInterval = int64(config.ReportInterval.ToDuration().Seconds())
	}
	if FlagPollInterval == 2 && config.PollInterval != 0 {
		FlagPollInterval = int64(config.PollInterval.ToDuration().Seconds())
	}
	if FlagCryptoKey == "" && config.CryptoKey != "" {
		FlagCryptoKey = config.CryptoKey
	}
	if FlagKey == "" && config.Key != "" {
		FlagKey = config.Key
	}
}

func readEnvVars() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if interval, err := strconv.ParseInt(envReportInterval, 10, 64); err == nil {
			FlagReportInterval = interval
		} else {
			zap.L().Error("Failed to parse REPORT_INTERVAL", zap.Error(err))
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if interval, err := strconv.ParseInt(envPollInterval, 10, 64); err == nil {
			FlagPollInterval = interval
		} else {
			zap.L().Error("Failed to parse POLL_INTERVAL", zap.Error(err))
		}
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		FlagCryptoKey = envCryptoKey
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		FlagKey = envKey
	}

	if envConfigFile := os.Getenv("CONFIG"); envConfigFile != "" {
		FlagConfigFile = envConfigFile
	}
}

func validateAndLogFlags() {
	if FlagReportInterval <= 0 {
		zap.L().Warn("Report interval must be positive, using default value",
			zap.Int64("default", 10))
		FlagReportInterval = 10
	}

	if FlagPollInterval <= 0 {
		zap.L().Warn("Poll interval must be positive, using default value",
			zap.Int64("default", 2))
		FlagPollInterval = 2
	}

	zap.L().Info(
		"Agent configuration",
		zap.String("address", FlagRunAddr),
		zap.Int64("report_interval", FlagReportInterval),
		zap.Int64("poll_interval", FlagPollInterval),
		zap.String("crypto_key", FlagCryptoKey),
		zap.String("key", config.MaskSensitive(FlagKey)),
		zap.String("config_file", FlagConfigFile),
	)
}
