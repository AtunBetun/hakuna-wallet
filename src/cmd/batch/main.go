package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/batch"
	"github.com/atunbetun/hakuna-wallet/pkg/db"
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

	// TODO; this has mutation, could be better
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	logger.Logger.Debug("configs parsed", zap.Any("cfg", cfg))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	databaseCfg := db.FromAppConfig(cfg)
	conn, err := db.Open(ctx, databaseCfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(conn); err != nil {
			panic(fmt.Sprintf("closing database connection: %s", err))
		}
	}()

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
