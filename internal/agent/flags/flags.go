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
//
// Поддерживаемые переменные окружения:
//
//	ADDRESS
//	REPORT_INTERVAL
//	POLL_INTERVAL
//	KEY
//	RATE_LIMIT
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
	if envKey := os.Getenv("KEY"); envKey != "" {
		zap.L().Info("KEY: ", zap.String("envKey", envKey))
		FlagKey = envKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		FlagRateLimit, err = strconv.ParseInt(envRateLimit, 10, 64)
		if err != nil {
			zap.L().Info("Error parse RATE_LIMIT", zap.Error(err))
		}
	}

	zap.L().Info(
		"Agent settings",
		zap.String("Server address", FlagRunAddr),
		zap.Int64("Report interval (sec)", FlagReportInterval),
		zap.Int64("Poll interval (sec)", FlagPollInterval),
		zap.String("Hash key", "[REDACTED]"), 
		zap.Int64("Rate limit", FlagRateLimit),
	)
}
