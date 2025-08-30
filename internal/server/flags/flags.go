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
	flag.StringVar(&FlagCryptoKey, "c", "", "path to file with private key for encryption")
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

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		FlagStoreInterval, err = strconv.ParseInt(envStoreInterval, 10, 64)
		if err != nil {
			zap.L().Info("Error parse STORE_INTERVAL", zap.Error(err))
		}
	}

	if envStorePath := os.Getenv("FILE_STORAGE_PATH"); envStorePath != "" {
		zap.L().Info("FILE_STORAGE_PATH: ", zap.String("envStorePath", envStorePath))
		FlagFileStoragePath = envStorePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		FlagRestore, err = strconv.ParseBool(envRestore)
		if err != nil {
			zap.L().Info("Error parse RESTORE", zap.Error(err))
		}
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		zap.L().Info("DATABASE_DSN: ", zap.String("envDatabaseDNS", envDatabaseDSN))
		FlagDatabaseDSN = envDatabaseDSN
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		zap.L().Info("KEY: ", zap.String("envKey", envKey))
		FlagKey = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		zap.L().Info("CRYPTO_KEY: ", zap.String("envCryptoKey", envCryptoKey))
		FlagCryptoKey = envCryptoKey
	}

	zap.L().Info(
		"Server settings",
		zap.String("Running server address: ", FlagRunAddr),
		zap.String("Running database address: ", FlagDatabaseDSN),
		zap.Int64("Store metrics interval: ", FlagStoreInterval),
		zap.String("Store path: ", FlagFileStoragePath),
		zap.Bool("Is restore: ", FlagRestore),
		zap.String("CryptoKey adress", FlagCryptoKey),
	)
}
