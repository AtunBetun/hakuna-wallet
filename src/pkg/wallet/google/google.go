package google

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
)

// Config carries the minimum attributes required to describe a Google Wallet pass for tests.
type Config struct {
	IssuerEmail string
	ClassID     string
	BinRange    string
	// ServiceAccountJSON holds the credential payload if required by a real integration.
	ServiceAccountJSON string
}

// Clock abstracts time retrieval to keep output deterministic in tests.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

// Generator produces lightweight JSON representations of Google Wallet passes that the CLI can persist.
type Generator struct {
	cfg   Config
	clock Clock
}

// Option returns a generator with the desired dependency overrides, keeping the original immutable.
type Option func(Generator) Generator

// NewGenerator builds a generator with the production clock.
func NewGenerator(cfg Config, opts ...Option) Generator {
	g := Generator{
		cfg:   cfg,
		clock: systemClock{},
	}

	for _, opt := range opts {
		g = opt(g)
	}

	return g
}

// WithClock swaps the clock implementation (useful to freeze time in tests).
func WithClock(clock Clock) Option {
	return func(g Generator) Generator {
		g.clock = clock
		return g
	}
}

// Generate renders a JSON artifact that callers can upload to Google Wallet or store as a fixture.
func (g Generator) Generate(_ context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
	if ticket.ID == "" {
		return wallet.Artifact{}, fmt.Errorf("ticket id is required")
	}
	if g.cfg.IssuerEmail == "" {
		return wallet.Artifact{}, fmt.Errorf("issuer email is required")
	}

	payload := map[string]any{
		"objectId": fmt.Sprintf("%s.%s", g.cfg.IssuerEmail, ticket.ID),
		"classId":  g.cfg.ClassID,
		"state":    "ACTIVE",
		"barcode":  ticket.Barcode,
		"description": map[string]string{
			"text": ticket.Description,
		},
		"holder": map[string]string{
			"name": ticket.FullName,
		},
		"event": map[string]string{
			"id":      ticket.EventID,
			"series":  ticket.EventSeriesID,
			"typeId":  ticket.TicketTypeID,
			"orderId": ticket.OrderID,
		},
		"meta": map[string]string{
			"issuedAt": g.clock.Now().Format(time.RFC3339Nano),
			"binRange": g.cfg.BinRange,
		},
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return wallet.Artifact{}, fmt.Errorf("could not marshal google wallet payload: %w", err)
	}

	return wallet.Artifact{
		Platform:    "google",
		FileName:    fmt.Sprintf("%s.json", ticket.ID),
		ContentType: "application/json",
		Data:        data,
	}, nil
}
