// Package flags предоставляет функциональность для обработки флагов командной строки
// и переменных окружения сервера метрик.
//
// Пакет поддерживает:
//   - Парсинг флагов командной строки
//   - Чтение значений из переменных окружения
//   - Приоритет переменных окружения над флагами
//   - Валидацию и логирование параметров
package flags

import (
	"flag"
	"os"
	"strconv"

	"github.com/MPoline/alert_service_yp/internal/config"
	"go.uber.org/zap"
)

// Глобальные переменные с параметрами сервера
var (
	// FlagRunAddr - адрес и порт для запуска сервера (флаг -a, переменная ADDRESS)
	FlagRunAddr string

	// FlagStoreInterval - интервал сохранения метрик в секундах (флаг -i, переменная STORE_INTERVAL)
	FlagStoreInterval int64

	// FlagFileStoragePath - путь к файлу для сохранения метрик (флаг -f, переменная FILE_STORAGE_PATH)
	FlagFileStoragePath string

	// FlagRestore - флаг восстановления метрик из файла при старте (флаг -r, переменная RESTORE)
	FlagRestore bool

	// FlagDatabaseDSN - строка подключения к БД (флаг -d, переменная DATABASE_DSN)
	FlagDatabaseDSN string

	// FlagKey - ключ для подписи данных (флаг -k, переменная KEY)
	FlagKey string

	// FlagCryptoKey - путь до файла с приватным ключом
	FlagCryptoKey string

	FlagConfigFile string

	// FlagTrustedSubnet - CIDR подсеть доверенных IP адресов (флаг -t, переменная TRUSTED_SUBNET)
	FlagTrustedSubnet string
)

// ParseFlags обрабатывает аргументы командной строки и переменные окружения.
// Приоритет значений: переменные окружения > флаги командной строки > значения по умолчанию.
//
// Поддерживаемые флаги:
//
//	-a : адрес сервера (по умолчанию ":8080")
//	-i : интервал сохранения в секундах (по умолчанию 300)
//	-f : путь к файлу метрик (по умолчанию "./savedMetrics")
//	-r : восстановить метрики из файла (по умолчанию false)
//	-d : строка подключения к БД (по умолчанию "")
//	-k : ключ для подписи (по умолчанию "+randomSrting+")
//	-с : ассиметричное шифрование (по умолчанию не используется)
//	-t : доверенная подсеть в формате CIDR (по умолчанию "")
//
// Пример использования:
//
//	flags.ParseFlags()
//	addr := flags.FlagRunAddr
func ParseFlags() {
	var err error
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&FlagStoreInterval, "i", 300, "frequency of save metrics")
	flag.StringVar(&FlagFileStoragePath, "f", "./savedMetrics", "address of file for save metrics")
	flag.BoolVar(&FlagRestore, "r", false, "read metrics from file")
	flag.StringVar(&FlagDatabaseDSN, "d", "", "address and port to run database")
	flag.StringVar(&FlagKey, "k", "+randomSrting+", "key hashSHA256")
	flag.StringVar(&FlagCryptoKey, "crypto-key", "", "path to file with private key for encryption")
	flag.StringVar(&FlagConfigFile, "config", "", "path to configuration file")
	flag.StringVar(&FlagConfigFile, "c", "", "path to configuration file (shorthand)")
	flag.StringVar(&FlagTrustedSubnet, "t", "", "trusted subnet in CIDR format")

	flag.Parse()

	var fileConfig *config.ServerConfig
	if FlagConfigFile != "" {
		fileConfig, err = config.LoadServerConfig(FlagConfigFile)
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

func applyFileConfig(config *config.ServerConfig) {
	if FlagRunAddr == ":8080" && config.Address != "" {
		FlagRunAddr = config.Address
	}
	if FlagStoreInterval == 300 && config.StoreInterval != 0 {
		FlagStoreInterval = int64(config.StoreInterval.ToDuration().Seconds())
	}
	if FlagFileStoragePath == "./savedMetrics" && config.FileStoragePath != "" {
		FlagFileStoragePath = config.FileStoragePath
	}
	if !FlagRestore && config.Restore {
		FlagRestore = config.Restore
	}
	if FlagDatabaseDSN == "" && config.DatabaseDSN != "" {
		FlagDatabaseDSN = config.DatabaseDSN
	}
	if FlagKey == "+randomSrting+" && config.Key != "" {
		FlagKey = config.Key
	}
	if FlagCryptoKey == "" && config.CryptoKey != "" {
		FlagCryptoKey = config.CryptoKey
	}

	if FlagTrustedSubnet == "" && config.TrustedSubnet != "" {
		FlagTrustedSubnet = config.TrustedSubnet
	}
}

func readEnvVars() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if interval, err := strconv.ParseInt(envStoreInterval, 10, 64); err == nil {
			FlagStoreInterval = interval
		} else {
			zap.L().Error("Failed to parse STORE_INTERVAL", zap.Error(err))
		}
	}

	if envStorePath := os.Getenv("FILE_STORAGE_PATH"); envStorePath != "" {
		FlagFileStoragePath = envStorePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if restore, err := strconv.ParseBool(envRestore); err == nil {
			FlagRestore = restore
		} else {
			zap.L().Error("Failed to parse RESTORE", zap.Error(err))
		}
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		FlagDatabaseDSN = envDatabaseDSN
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		FlagKey = envKey
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		FlagCryptoKey = envCryptoKey
	}

	if envConfigFile := os.Getenv("CONFIG"); envConfigFile != "" {
		FlagConfigFile = envConfigFile
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		FlagTrustedSubnet = envTrustedSubnet
	}
}

func validateAndLogFlags() {
	if FlagStoreInterval < 0 {
		zap.L().Warn("Store interval cannot be negative, using default value",
			zap.Int64("default", 300))
		FlagStoreInterval = 300
	}

	zap.L().Info(
		"Server configuration",
		zap.String("address", FlagRunAddr),
		zap.Int64("store_interval", FlagStoreInterval),
		zap.String("file_storage_path", FlagFileStoragePath),
		zap.Bool("restore", FlagRestore),
		zap.String("database_dsn", FlagDatabaseDSN),
		zap.String("key", config.MaskSensitive(FlagKey)),
		zap.String("crypto_key", FlagCryptoKey),
		zap.String("config_file", FlagConfigFile),
		zap.String("trusted_subnet", FlagTrustedSubnet), 
	)
}