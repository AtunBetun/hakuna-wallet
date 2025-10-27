package batch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/aws"
	"github.com/atunbetun/hakuna-wallet/pkg/db"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ticketFetcher func(ctx context.Context, cfg tickets.TicketTailorConfig) ([]tickets.TTIssuedTicket, error)

type passGenerator func(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error)

// returns file path and error
type artifactSink func(ctx context.Context, artifact wallet.Artifact) (string, error)

type Platform string

// TODO: this probably does nothing
const defaultTicketStatus = "issued"
const (
	PlatformApple Platform = "apple"
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
	TicketID         string
	Platform         Platform
	FileName         string
	Email            string
	FullArtifactPath string
}

// walletTicketSyncer orchestrates fetching tickets, generating wallet passes, and persisting artifacts.
type walletTicketSyncer struct {
	ticketConfig   tickets.TicketTailorConfig `validate:"required"`
	TicketFetcher  ticketFetcher              `validate:"required"`
	AppleGenerator passGenerator              `validate:"required"`
	ArtifactSink   artifactSink               `validate:"required"`
	TicketStatus   string                     `validate:"required"`
	AppConfig      pkg.AppConfig              `validate:"required"`
	DB             *gorm.DB                   `validate:"-"`
	S3Client       *aws.S3Client              `validate:"required"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

// NewWalletTicketSyncer wires default dependencies based on the provided configuration.
func NewWalletTicketSyncer(ctx context.Context, cfg pkg.AppConfig, conn *gorm.DB) (*walletTicketSyncer, error) {
	ticketCfg, err := tickets.NewTicketTailorConfig(cfg)
	if err != nil {
		return nil, err
	}

	sink, err := newFileSink(cfg.TicketsDir)
	if err != nil {
		return nil, err
	}

	appleGen, err := newAppleGenerator(cfg, EmbeddedAppleGenerator)
	if err != nil {
		return nil, err
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return &walletTicketSyncer{}, err
	}

	s3, err := aws.NewS3Client(cfg.S3Bucket, &awsConfig)
	if err != nil {
		return &walletTicketSyncer{}, err
	}

	out := &walletTicketSyncer{
		ticketConfig:   ticketCfg,
		TicketFetcher:  newTicketTailorTicketFetcher(), // TODO: remove this
		AppleGenerator: appleGen,
		ArtifactSink:   sink,
		TicketStatus:   defaultTicketStatus,
		DB:             conn,
		AppConfig:      cfg,
		S3Client:       s3,
	}
	err = validate.Struct(out)
	if err != nil {
		return &walletTicketSyncer{}, err
	}
	return out, err
}

// SyncTickets executes the ticket ingestion flow, returning a summary of created artifacts.
func (g *walletTicketSyncer) SyncTickets(ctx context.Context) (GenerationSummary, error) {
	if g.TicketFetcher == nil {
		return GenerationSummary{}, fmt.Errorf("ticket fetcher is not configured")
	}
	if g.ArtifactSink == nil {
		return GenerationSummary{}, fmt.Errorf("artifact sink is not configured")
	}

	logger.Logger.Debug(
		"Fetching issued tickets",
		zap.String("event_id", g.ticketConfig.EventId),
	)
	ticketsBatch, err := g.TicketFetcher(ctx, g.ticketConfig)
	if err != nil {
		return GenerationSummary{}, fmt.Errorf("fetching ticket tailor issued tickets: %w", err)
	}
	logger.Logger.Debug(
		"Fetched tickets batch",
		zap.String("event_id", g.ticketConfig.EventId),
		zap.Int("count", len(ticketsBatch)),
	)

	currentTickets, err := db.GetProducedPasses(ctx, g.DB, db.AppleWalletChannel)
	if err != nil {
		return GenerationSummary{}, fmt.Errorf("getting produced passes: %w", err)
	}

	tickets := ticketsForSync(ticketsBatch, currentTickets)
	logger.Logger.Info(
		"Generating tickets",
		zap.Int("count", len(tickets)),
	)
	created, err := g.generateTickets(ctx, tickets)
	if err != nil {
		return GenerationSummary{}, fmt.Errorf("generating tickets: %w", err)
	}

	// _ = mailer.NewAppleMailDialer(
	// 	"smtp.mail.me.com",
	// 	587,
	// 	"adesaintmalo@icloud.com",
	// 	g.AppConfig.ApplePassword,
	// )

	// TODO: this should be from the created, not from the tickets

	logger.Logger.Debug(
		"Fetched tickets batch",
		zap.String("event_id", g.ticketConfig.EventId),
		zap.Int("count", len(ticketsBatch)),
	)
	errs := []error{}
	for _, v := range created {
		err = g.processTicket(ctx, v)
		if err != nil {
			logger.Logger.Error("processing tickeet", zap.Error(err))
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return GenerationSummary{}, errors.Join(errs...)
	}

	return GenerationSummary{Artifacts: created}, nil
}

func (g *walletTicketSyncer) processTicket(
	ctx context.Context,
	artifact GeneratedArtifact,
) error {
	logger.Logger.Debug(
		"Marking ticket as ticket created",
	)
	err := db.SetPassProduced(
		ctx,
		g.DB,
		db.AppleWalletChannel,
		artifact.TicketID,
		artifact.Email,
		time.Now(),
	)
	if err != nil {
		logger.Logger.Fatal("setting pass produced: %w", zap.Any("", err))
		return err
	}

	logger.Logger.Debug(
		"Uploading to s3",
		zap.String("event_id", g.ticketConfig.EventId),
	)
	_, err = g.S3Client.UploadFile(
		ctx,
		g.AppConfig.S3Bucket,
		ticketKey(artifact.FileName),
		artifact.FullArtifactPath,
	)
	if err != nil {
		logger.Logger.Fatal("uploading file: %w", zap.Any("", err))
		return err
	}
	_, err = g.S3Client.PresignURLDefault(
		ctx,
		g.AppConfig.S3Bucket,
		ticketKey(artifact.FileName),
	)
	if err != nil {
		logger.Logger.Fatal("presign url: %w", zap.Any("", err))
		return err
	}
	return nil
}
func ticketKey(ticketName string) string {
	key := "ham-2026/apple-wallet/" + ticketName
	return key
}

func ticketsForSync(
	ticketsBatch []tickets.TTIssuedTicket,
	currentTickets map[string]db.PassRecord,
) []tickets.TTIssuedTicket {

	var missing []tickets.TTIssuedTicket
	for _, ticket := range ticketsBatch {
		if _, exists := currentTickets[ticket.ID]; !exists {
			missing = append(missing, ticket)
		}
	}
	return missing
}

func (g *walletTicketSyncer) generateTickets(
	ctx context.Context,
	ticketsBatch []tickets.TTIssuedTicket,
) (
	[]GeneratedArtifact,
	error,
) {
	var created []GeneratedArtifact
	for _, tt := range ticketsBatch {
		logger.Logger.Debug(
			"Generating wallet passes for ticket",
			zap.Any("ticket", tt),
		)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if g.AppleGenerator != nil {
			info, err := g.generateAndPersist(ctx, g.AppleGenerator, tt, PlatformApple)
			if err != nil {
				return nil, err
			}
			created = append(created, info)
		}
	}
	return created, nil
}

func (g *walletTicketSyncer) generateAndPersist(
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

	fullPath, err := g.ArtifactSink(ctx, artifact)
	if err != nil {
		return GeneratedArtifact{}, fmt.Errorf("persisting %s wallet artifact for ticket %s: %w", platform, ticket.ID, err)
	}

	logger.Logger.Debug(
		"Persisted wallet artifact",
		zap.String("ticket_id", ticket.ID),
		zap.String("platform", string(platform)),
		zap.String("file_name", artifact.FileName),
	)
	return GeneratedArtifact{
		TicketID:         ticket.ID,
		Platform:         platform,
		FileName:         artifact.FileName,
		Email:            ticket.Email,
		FullArtifactPath: fullPath,
	}, nil
}

// TODO: remove this
func newTicketTailorTicketFetcher() ticketFetcher {
	return func(ctx context.Context, cfg tickets.TicketTailorConfig) ([]tickets.TTIssuedTicket, error) {
		return tickets.FetchAllIssuedTickets(ctx, cfg, tickets.Valid)
	}
}
