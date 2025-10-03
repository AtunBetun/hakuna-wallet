package main

import (
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
)

func main() {
	logger.Init()
	defer logger.Logger.Sync()
	logger.Logger.Info("Started")
}
