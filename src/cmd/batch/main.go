package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/apple_wallet"
	"github.com/atunbetun/hakuna-wallet/pkg/batch"
	"github.com/atunbetun/hakuna-wallet/pkg/google_wallet"
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
	batch.BatchProcess(ctx, cfg)

	// Create Apple Wallet ticket
	passData, err := apple_wallet.CreateAppleWalletTicket(ctx)
	if err != nil {
		logger.Logger.Fatal("Failed to create apple wallet ticket", zap.Error(err))
	}

	// Save the pass to a file
	err = os.WriteFile("hakuna.pkpass", passData, 0644)
	if err != nil {
		logger.Logger.Fatal("Failed to write pass to file", zap.Error(err))
	}

	logger.Logger.Info("Apple Wallet ticket created successfully")

	// Create Google Wallet ticket
	googleWalletService, err := google_wallet.NewGoogleWalletService(&cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to create google wallet service", zap.Error(err))
	}

	classId := fmt.Sprintf("%s.%s", cfg.GoogleIssuerEmail, "hakuna_matata_class")
	_, err = googleWalletService.CreateEventTicketClass(classId)
	if err != nil {
		logger.Logger.Fatal("Failed to create google wallet class", zap.Error(err))
	}

	objectId := fmt.Sprintf("%s.%s", cfg.GoogleIssuerEmail, "hakuna_matata_object")
	_, err = googleWalletService.CreateEventTicketObject(classId, objectId)
	if err != nil {
		logger.Logger.Fatal("Failed to create google wallet object", zap.Error(err))
	}

	logger.Logger.Info("Google Wallet ticket created successfully")
}
