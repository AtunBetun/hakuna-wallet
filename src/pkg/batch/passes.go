package batch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet/apple"
	"go.uber.org/zap"
)

type AppleGeneratorType string

const (
	EmbeddedAppleGenerator AppleGeneratorType = "embedded"
	DefaultAppleGenerator  AppleGeneratorType = "default"
)

func newAppleGenerator(cfg pkg.AppConfig, genType AppleGeneratorType) (passGenerator, error) {
	appleConfig, err := getAppleConfig(cfg)
	if err != nil {
		return nil, err
	}

	switch genType {
	case EmbeddedAppleGenerator:
		return func(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
			creator := apple.NewEmbeddedApplePassCreator(appleConfig)
			return creator.Create(ctx, ticket)
		}, nil
	case DefaultAppleGenerator:
		return func(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
			creator := apple.NewDefaultApplePassCreator(appleConfig)
			return creator.Create(ctx, ticket)
		}, nil
	default:
		return nil, fmt.Errorf("unknown apple generator type: %s", genType)
	}

}

func newFileSink(root string) (artifactSink, error) {
	if root == "" {
		return nil, fmt.Errorf("tickets dir cannot be empty")
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("creating tickets dir: %w", err)
	}

	// returns file path and error
	return func(ctx context.Context, artifact wallet.Artifact) (string, error) {
		if artifact.FileName == "" {
			return "", fmt.Errorf("artifact filename is required")
		}

		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		dir := filepath.Join(root, artifact.Platform)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("creating platform directory: %w", err)
		}

		fullPath := filepath.Join(dir, artifact.FileName)
		if err := os.WriteFile(fullPath, artifact.Data, 0o600); err != nil {
			return "", fmt.Errorf("writing artifact to %s: %w", fullPath, err)
		}
		logger.Logger.Debug(
			"Wrote wallet artifact to disk",
			zap.String("platform", artifact.Platform),
			zap.String("file_name", artifact.FileName),
			zap.String("path", fullPath),
		)
		return fullPath, nil
	}, nil
}

func getAppleConfig(cfg pkg.AppConfig) (apple.AppleConfig, error) {
	if cfg.ApplePassTypeID == "" {
		return apple.AppleConfig{}, fmt.Errorf("apple pass type identifier is required")
	}
	if cfg.AppleTeamID == "" {
		return apple.AppleConfig{}, fmt.Errorf("apple team identifier is required")
	}
	if cfg.AppleP12Path == "" {
		return apple.AppleConfig{}, fmt.Errorf("apple signing certificate path is required")
	}

	appleConfig := apple.AppleConfig{
		PassTypeIdentifier:         cfg.ApplePassTypeID,
		TeamIdentifier:             cfg.AppleTeamID,
		OrganizationName:           "Hakuna Wallet",
		Description:                "Hakuna Wallet Ticket",
		LogoText:                   "Hakuna Wallet",
		SigningCertificatePath:     cfg.AppleP12Path,
		SigningCertificatePassword: cfg.AppleP12Password,
		AppleRootCertificatePath:   cfg.AppleRootCertPath,
	}
	return appleConfig, nil
}
