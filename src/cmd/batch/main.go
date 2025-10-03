package main

import (
	"context"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/batch"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	defer logger.Logger.Sync()
	logger.Logger.Info("Started")

	err := godotenv.Load()
	if err != nil {
		logger.Logger.Fatal("Error loading .env file", zap.Any("err", err))
	}

	cfg := pkg.Config{}

	if err := env.Parse(&cfg); err != nil {
		logger.Logger.Fatal("Failed to parse env", zap.Any("err", err))
	}
	logger.Logger.Debug("configs parsed", zap.Any("cfg", cfg))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	batch.PurgeTickets(ctx, cfg)
	logger.Logger.Info("Success")

}
