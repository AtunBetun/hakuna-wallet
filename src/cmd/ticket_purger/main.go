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

	if pkg.ShouldLoadDotenv() {
		logger.Logger.Info("Loading .env")
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
	}

	cfg := pkg.AppConfig{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	logger.Logger.Debug("configs parsed", zap.Any("cfg", cfg))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err := batch.PurgeTickets(ctx, cfg)
	if err != nil {
		panic(err)
	}
	logger.Logger.Info("Success")

}
