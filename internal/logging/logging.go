// Package logging предоставляет централизованную инициализацию логгера zap.
//
// Пакет содержит:
// - Глобальный экземпляр логгера
// - Функции инициализации и синхронизации
//
// Рекомендуемый способ использования:
//
//	logger, err := logging.InitLog()
//	if err != nil {
//	    // обработка ошибки
//	}
//	defer logging.Sync()
package logging

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger - глобальный экземпляр логгера.
// Инициализируется при вызове InitLog().
var Logger *zap.Logger

// InitLog инициализирует логгер в development-режиме.
//
// Development-режим включает:
// - Логи в удобочитаемом формате
// - Stacktrace для сообщений уровня Error и выше
// - Логирование в stderr
//
// Возвращает:
//   - *zap.Logger: инициализированный логгер
//   - error: ошибка инициализации
//
// В случае ошибки инициализации вызывает panic.
func InitLog() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Logger initialization error", err)
		panic(err)
	}
	Logger = logger
	return logger, nil
}

// Sync синхронизирует буферизованные логи.
// Рекомендуется вызывать через defer при инициализации приложения.
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
