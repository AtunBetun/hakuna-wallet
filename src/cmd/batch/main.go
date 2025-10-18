package main

import (
	"context"
	"os"
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

	if shouldLoadDotenv() {
		logger.Logger.Info("Loading .env")
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
	}

	cfg := pkg.Config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	cfg, err := pkg.InitializeCertificates(cfg)
	if err != nil {
		panic(err)
	}

	logger.Logger.Debug("configs parsed", zap.Any("cfg", cfg))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	err = batch.GenerateTickets(ctx, cfg)
	if err != nil {
		panic(err)
	}
	logger.Logger.Info("Success")

}

func shouldLoadDotenv() bool {
	env, found := os.LookupEnv("APP_ENV")
	if found && env == "prod" {
		logger.Logger.Info("Not loading .env")
		return false
	}
	return true
}
