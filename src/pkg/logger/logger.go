package logger

import "go.uber.org/zap"

var Logger *zap.Logger

func Init() {
	l, _ := zap.NewProduction()
	Logger = l
}
