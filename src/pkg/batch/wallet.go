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
	"github.com/atunbetun/hakuna-wallet/pkg/wallet/google"
	"go.uber.org/zap"
)

const defaultTicketStatus = "issued"

type ticketFetcher func(ctx context.Context, cfg tickets.TicketTailorConfig, status string) ([]tickets.TTIssuedTicket, error)

type passGenerator func(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error)

type artifactSink func(ctx context.Context, artifact wallet.Artifact) error

type Platform string

const (
	PlatformApple  Platform = "apple"
	PlatformGoogle Platform = "google"
)

// TicketGenerator produces wallet artifacts from Ticket Tailor issued tickets.
type TicketGenerator interface {
	GenerateTickets(ctx context.Context) (GenerationSummary, error)
}

// GenerationSummary provides metadata about the artifacts created during a run.
type GenerationSummary struct {
	Artifacts []GeneratedArtifact
}

// GeneratedArtifact captures the origin of a created wallet artifact.
type GeneratedArtifact struct {
	TicketID string
	Platform Platform
	FileName string
}

// WalletTicketGenerator orchestrates fetching tickets, generating wallet passes, and persisting artifacts.
type WalletTicketGenerator struct {
	ticketConfig    tickets.TicketTailorConfig
	TicketFetcher   ticketFetcher
	AppleGenerator  passGenerator
	GoogleGenerator passGenerator
	ArtifactSink    artifactSink
	TicketStatus    string
}

// NewWalletTicketGenerator wires default dependencies based on the provided configuration.
func NewWalletTicketGenerator(cfg pkg.Config) (*WalletTicketGenerator, error) {
	ticketCfg, err := tickets.NewTicketTailorConfig(cfg)
	if err != nil {
		return nil, err
	}

	sink, err := prepareArtifactSink(cfg)
	if err != nil {
		return nil, err
	}

	appleGen, err := appleGenerator(cfg)
	if err != nil {
		return nil, err
	}

	googleGen, err := googleGenerator(cfg)
	if err != nil {
		return nil, err
	}

	return &WalletTicketGenerator{
		ticketConfig:    ticketCfg,
		TicketFetcher:   defaultIssuedTicketFetcher(), // TODO: remove this
		AppleGenerator:  appleGen,
		GoogleGenerator: googleGen,
		ArtifactSink:    sink,
		TicketStatus:    defaultTicketStatus,
	}, nil
}

// GenerateTickets executes the ticket ingestion flow, returning a summary of created artifacts.
func (g *WalletTicketGenerator) GenerateTickets(ctx context.Context) (GenerationSummary, error) {
	if g.TicketFetcher == nil {
		return GenerationSummary{}, fmt.Errorf("ticket fetcher is not configured")
	}
	if g.ArtifactSink == nil {
		return GenerationSummary{}, fmt.Errorf("artifact sink is not configured")
	}

	status := g.TicketStatus
	if status == "" {
		status = defaultTicketStatus
	}

	logger.Logger.Debug(
		"Fetching issued tickets",
		zap.String("event_id", g.ticketConfig.EventId),
		zap.String("status", status),
	)
	ticketsBatch, err := g.TicketFetcher(ctx, g.ticketConfig, status)
	if err != nil {
		return GenerationSummary{}, fmt.Errorf("fetching ticket tailor issued tickets: %w", err)
	}

	logger.Logger.Debug(
		"Fetched issued tickets batch",
		zap.String("event_id", g.ticketConfig.EventId),
		zap.Int("count", len(ticketsBatch)),
	)
	var created []GeneratedArtifact

	for _, tt := range ticketsBatch {
		logger.Logger.Debug(
			"Generating wallet passes for ticket",
			zap.Any("ticket", tt),
		)
		if ctx.Err() != nil {
			return GenerationSummary{}, ctx.Err()
		}

		if g.AppleGenerator != nil {
			info, err := g.generateAndPersist(ctx, g.AppleGenerator, tt, PlatformApple)
			if err != nil {
				return GenerationSummary{}, err
			}
			created = append(created, info)
		}

		if g.GoogleGenerator != nil {
			info, err := g.generateAndPersist(ctx, g.GoogleGenerator, tt, PlatformGoogle)
			if err != nil {
				return GenerationSummary{}, err
			}
			created = append(created, info)
		}
	}

	return GenerationSummary{Artifacts: created}, nil
}

func (g *WalletTicketGenerator) generateAndPersist(
	ctx context.Context,
	generator passGenerator,
	ticket tickets.TTIssuedTicket,
	platform Platform,
) (GeneratedArtifact, error) {
	logger.Logger.Debug(
		"Starting wallet artifact generation",
		zap.String("ticket_id", ticket.ID),
		zap.String("platform", string(platform)),
	)
	artifact, err := generator(ctx, ticket)
	if err != nil {
		return GeneratedArtifact{}, fmt.Errorf("generating %s wallet pass for ticket %s: %w", platform, ticket.ID, err)
	}

	if err := g.ArtifactSink(ctx, artifact); err != nil {
		return GeneratedArtifact{}, fmt.Errorf("persisting %s wallet artifact for ticket %s: %w", platform, ticket.ID, err)
	}

	logger.Logger.Debug(
		"Persisted wallet artifact",
		zap.String("ticket_id", ticket.ID),
		zap.String("platform", string(platform)),
		zap.String("file_name", artifact.FileName),
	)
	return GeneratedArtifact{
		TicketID: ticket.ID,
		Platform: platform,
		FileName: artifact.FileName,
	}, nil
}

// TODO: remove this
func defaultIssuedTicketFetcher() ticketFetcher {
	return func(ctx context.Context, cfg tickets.TicketTailorConfig, status string) ([]tickets.TTIssuedTicket, error) {
		return tickets.FetchAllIssuedTickets(ctx, cfg, tickets.Valid)
	}
}

func appleGenerator(cfg pkg.Config) (passGenerator, error) {
	if cfg.ApplePassTypeID == "" {
		return nil, fmt.Errorf("apple pass type identifier is required")
	}
	if cfg.AppleTeamID == "" {
		return nil, fmt.Errorf("apple team identifier is required")
	}
	if cfg.AppleP12Path == "" {
		return nil, fmt.Errorf("apple signing certificate path is required")
	}

	appleConfig := apple.AppleConfig{
		PassTypeIdentifier:         cfg.ApplePassTypeID,
		TeamIdentifier:             cfg.AppleTeamID,
		OrganizationName:           "Hakuna Wallet",
		Description:                "Hakuna Wallet Ticket",
		LogoText:                   "Hakuna Wallet",
		SigningCertificatePath:     cfg.AppleP12Path,
		SigningCertificatePassword: cfg.AppleP12Password,
		AppleRootCertificatePath:   cfg.AppleRootCert,
	}

	creator := apple.NewApplePassCreator(appleConfig)

	return func(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
		return creator.Create(ctx, ticket)
	}, nil
}

func googleGenerator(cfg pkg.Config) (passGenerator, error) {
	if cfg.GoogleIssuerEmail == "" {
		return nil, fmt.Errorf("google issuer email is required")
	}
	if cfg.GoogleServiceAccountJSON == "" {
		return nil, fmt.Errorf("google service account json is required")
	}

	generator := google.NewGenerator(google.Config{
		IssuerEmail:        cfg.GoogleIssuerEmail,
		ClassID:            "hakuna.wallet.ticket", // TODO: inject
		BinRange:           "000000-999999",
		ServiceAccountJSON: cfg.GoogleServiceAccountJSON,
	})

	return func(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
		return generator.Generate(ctx, ticket)
	}, nil
}

func prepareArtifactSink(cfg pkg.Config) (artifactSink, error) {
	return newFileSink(cfg.TicketsDir)
}

func newFileSink(root string) (artifactSink, error) {
	if root == "" {
		return nil, fmt.Errorf("tickets dir cannot be empty")
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("creating tickets dir: %w", err)
	}

	return func(ctx context.Context, artifact wallet.Artifact) error {
		if artifact.FileName == "" {
			return fmt.Errorf("artifact filename is required")
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		dir := filepath.Join(root, artifact.Platform)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating platform directory: %w", err)
		}

		fullPath := filepath.Join(dir, artifact.FileName)
		if err := os.WriteFile(fullPath, artifact.Data, 0o600); err != nil {
			return fmt.Errorf("writing artifact to %s: %w", fullPath, err)
		}
		logger.Logger.Debug(
			"Wrote wallet artifact to disk",
			zap.String("platform", artifact.Platform),
			zap.String("file_name", artifact.FileName),
			zap.String("path", fullPath),
		)
		return nil
	}, nil
}
