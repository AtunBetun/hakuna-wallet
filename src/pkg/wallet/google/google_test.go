package google_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet/google"
)

type fakeClock struct {
	t time.Time
}

func (f fakeClock) Now() time.Time {
	return f.t
}

func TestGeneratorProduceDeterministicArtifact(t *testing.T) {
	t.Helper()

	issuedAt := time.Date(2024, time.July, 1, 12, 0, 0, 0, time.UTC)
	gen := google.NewGenerator(
		google.Config{
			IssuerEmail:        "issuer@hakuna.dev",
			ClassID:            "hakuna.pass.class",
			BinRange:           "400000-499999",
			ServiceAccountJSON: "{}",
		},
		google.WithClock(fakeClock{t: issuedAt}),
	)

	ticket := tickets.TTIssuedTicket{
		ID:            "tt_123",
		Description:   "Hakuna Integration Gala",
		FullName:      "Nala Hakuna",
		Barcode:       "BR-123",
		EventID:       "ev_123",
		EventSeriesID: "series_123",
		TicketTypeID:  "vip",
		OrderID:       "order_123",
	}

	artifact, err := gen.Generate(context.Background(), ticket)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if artifact.Platform != "google" {
		t.Fatalf("expected platform google, got %s", artifact.Platform)
	}
	if artifact.FileName != "tt_123.json" {
		t.Fatalf("expected filename tt_123.json, got %s", artifact.FileName)
	}
	if artifact.ContentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", artifact.ContentType)
	}

	var payload map[string]any
	if err := json.Unmarshal(artifact.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload["objectId"] != "issuer@hakuna.dev.tt_123" {
		t.Errorf("unexpected objectId: %v", payload["objectId"])
	}
	if payload["classId"] != "hakuna.pass.class" {
		t.Errorf("unexpected classId: %v", payload["classId"])
	}
	if payload["state"] != "ACTIVE" {
		t.Errorf("unexpected state: %v", payload["state"])
	}

	meta, ok := payload["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta field missing or not an object: %v", payload["meta"])
	}
	if meta["issuedAt"] != issuedAt.Format(time.RFC3339Nano) {
		t.Errorf("unexpected issuedAt: %v", meta["issuedAt"])
	}
	if meta["binRange"] != "400000-499999" {
		t.Errorf("unexpected binRange: %v", meta["binRange"])
	}
}
