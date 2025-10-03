package logger

import "go.uber.org/zap"

var Logger *zap.Logger

func Init() {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}

	// Enable debug level logs
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	// Build the logger
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	Logger = l
}
