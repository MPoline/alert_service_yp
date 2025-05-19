package logging

import (
	"fmt"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLog() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Logger initialization error", err)
		panic(err)
	}
	return logger, nil
}

func Sync() {
	_ = Logger.Sync()
}
