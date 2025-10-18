package batch_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/batch"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
)

func TestGenerateWalletTickets(t *testing.T) {
	logger.Init()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	issued := []tickets.TTIssuedTicket{
		stubTicket("tt_100", "Nala", "BR-HK-100"),
		stubTicket("tt_101", "Simba", "BR-HK-101"),
	}

	cfg := pkg.Config{
		TicketTailorAPIKey:  "test-api-key",
		TicketTailorEventId: "4242",
		TicketTailorBaseUrl: "https://example.invalid",
		AppleP12Path:        "testdata/certificates/hakuna-test.p12",
		AppleP12Password:    "password",
		AppleRootCertPath:   "testdata/certificates/apple-root.cer",
		ApplePassTypeID:     "pass.com.hakuna.integration",
		AppleTeamID:         "TEAMSAMPLE",
		TicketsDir:          t.TempDir(),
	}

	generator, err := batch.NewWalletTicketGenerator(cfg)
	if err != nil {
		t.Fatalf("generator: %v", err)
	}

	expectedStatus := "ready"
	generator.TicketStatus = expectedStatus

	fetchCalls := 0
	generator.TicketFetcher = func(ctx context.Context, _ tickets.TicketTailorConfig, status string) ([]tickets.TTIssuedTicket, error) {
		fetchCalls++
		if status != expectedStatus {
			return nil, fmt.Errorf("expected status %q, got %q", expectedStatus, status)
		}
		return issued, nil
	}

	expected := map[batch.Platform]map[string]string{
		batch.PlatformApple: {},
	}

	generator.AppleGenerator = func(_ context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
		payload := fmt.Sprintf("apple:%s", ticket.ID)
		fileName := fmt.Sprintf("%s-apple.mock", ticket.ID)
		expected[batch.PlatformApple][fileName] = payload
		return wallet.Artifact{
			Platform:    string(batch.PlatformApple),
			FileName:    fileName,
			ContentType: "application/json",
			Data:        []byte(payload),
		}, nil
	}

	saved := map[batch.Platform]map[string][]byte{
		batch.PlatformApple: {},
	}

	generator.ArtifactSink = func(_ context.Context, artifact wallet.Artifact) error {
		platform := batch.Platform(artifact.Platform)
		if saved[platform] == nil {
			saved[platform] = map[string][]byte{}
		}
		saved[platform][artifact.FileName] = append([]byte(nil), artifact.Data...)
		return nil
	}

	summary, err := generator.GenerateTickets(ctx)
	if err != nil {
		t.Fatalf("generate wallet tickets: %v", err)
	}

	if fetchCalls != 1 {
		t.Fatalf("expected 1 fetch call, got %d", fetchCalls)
	}

	expectedCount := len(issued) * len(expected)
	if len(summary.Artifacts) != expectedCount {
		t.Fatalf("expected %d artifacts in summary, got %d", expectedCount, len(summary.Artifacts))
	}

	for platform, artifacts := range expected {
		savedArtifacts, ok := saved[platform]
		if !ok {
			t.Fatalf("no artifacts stored for platform %s", platform)
		}
		if len(artifacts) != len(savedArtifacts) {
			t.Fatalf("expected %d artifacts for %s, got %d", len(artifacts), platform, len(savedArtifacts))
		}
		for file, payload := range artifacts {
			data, ok := savedArtifacts[file]
			if !ok {
				t.Fatalf("missing artifact %s for %s", file, platform)
			}
			if string(data) != payload {
				t.Fatalf("artifact %s payload mismatch: want %q got %q", file, payload, data)
			}
		}
	}

	seen := map[string]struct{}{}
	for _, artifact := range summary.Artifacts {
		key := fmt.Sprintf("%s/%s", artifact.TicketID, artifact.Platform)
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("duplicate artifact summary entry for %s", key)
		}
		seen[key] = struct{}{}
		if artifact.FileName == "" {
			t.Fatalf("artifact summary missing filename for ticket %s", artifact.TicketID)
		}
		if _, ok := expected[artifact.Platform][artifact.FileName]; !ok {
			t.Fatalf("unexpected artifact %s for platform %s in summary", artifact.FileName, artifact.Platform)
		}
	}
}

func stubTicket(id string, name string, barcode string) tickets.TTIssuedTicket {
	return tickets.TTIssuedTicket{
		ID:          id,
		Description: "Hakuna Sunset",
		FullName:    name,
		Barcode:     barcode,
		Status:      "issued",
	}
}
