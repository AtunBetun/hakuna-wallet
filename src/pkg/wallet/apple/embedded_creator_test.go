package apple

import (
	"context"
	"strings"
	"testing"

	"github.com/alvinbaena/passkit"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
)

func TestEmbeddedCreatorBuildsPassFromTemplate(t *testing.T) {
	logger.Init()
	t.Helper()

	signer := &capturingSigner{}
	signingLoader := func(_, _, _ string) (*passkit.SigningInformation, error) {
		return &passkit.SigningInformation{}, nil
	}

	creator := NewEmbeddedApplePassCreator(
		AppleConfig{
			PassTypeIdentifier:         "pass.com.hakuna.integration",
			TeamIdentifier:             "TEAMHAKUNA",
			OrganizationName:           "Hakuna Wallet",
			Description:                "Hakuna Wallet Ticket",
			LogoText:                   "Hakuna",
			SigningCertificatePath:     "/tmp/cert.p12",
			SigningCertificatePassword: "integration-password",
			AppleRootCertificatePath:   "/tmp/root.cer",
		},
	)
	creator.Signer = signer
	creator.SigningInfoLoader = signingLoader

	ticket := tickets.TTIssuedTicket{
		ID:       "tt_embed_001",
		Barcode:  "EMBED-QR-001",
		FullName: "Nala Hakuna",
	}

	artifact, err := creator.Create(context.Background(), ticket)
	if err != nil {
		t.Fatalf("generate embedded pass: %v", err)
	}

	if artifact.Platform != "apple" {
		t.Fatalf("expected platform apple, got %q", artifact.Platform)
	}
	if artifact.FileName != "tt_embed_001.pkpass" {
		t.Fatalf("unexpected filename %q", artifact.FileName)
	}
	if artifact.ContentType != "application/vnd.apple.pkpass" {
		t.Fatalf("unexpected content type %q", artifact.ContentType)
	}

	if signer.callCount != 1 {
		t.Fatalf("expected signer invoked once, got %d", signer.callCount)
	}
	if signer.pass == nil {
		t.Fatalf("signer did not capture pass")
	}

	pass := signer.pass
	if pass.SerialNumber != "tt_embed_001" {
		t.Fatalf("unexpected serial number %q", pass.SerialNumber)
	}
	if pass.PassTypeIdentifier != "pass.com.hakuna.integration" {
		t.Fatalf("pass type identifier mismatch: %q", pass.PassTypeIdentifier)
	}
	if pass.TeamIdentifier != "TEAMHAKUNA" {
		t.Fatalf("team identifier mismatch: %q", pass.TeamIdentifier)
	}
	if pass.OrganizationName != "Hakuna Wallet" {
		t.Fatalf("organization name mismatch: %q", pass.OrganizationName)
	}
	if pass.Description != "Hakuna Wallet Ticket" {
		t.Fatalf("pass description mismatch: %q", pass.Description)
	}

	if len(pass.Barcodes) != 1 {
		t.Fatalf("expected one barcode, got %d", len(pass.Barcodes))
	}
	if pass.Barcodes[0].Message != "EMBED-QR-001" {
		t.Fatalf("barcode payload mismatch: %q", pass.Barcodes[0].Message)
	}
	if pass.Barcodes[0].AltText != "EMBED-QR-001" {
		t.Fatalf("barcode alt text mismatch: %q", pass.Barcodes[0].AltText)
	}
	if pass.Barcodes[0].MessageEncoding != "iso-8859-1" {
		t.Fatalf("unexpected barcode encoding: %q", pass.Barcodes[0].MessageEncoding)
	}

	if pass.BoardingPass == nil {
		t.Fatalf("expected boarding pass section to be present")
	}

	foundPassenger := false
	for _, field := range pass.BoardingPass.SecondaryFields {
		if strings.EqualFold(field.Label, "passenger") || strings.EqualFold(field.Key, "passenger") {
			foundPassenger = true
			if field.Value != "Nala Hakuna" {
				t.Fatalf("passenger field mismatch: %+v", field)
			}
			break
		}
	}
	if !foundPassenger {
		t.Fatalf("passenger field missing; secondary fields: %+v", pass.BoardingPass.SecondaryFields)
	}

	memTemplate, ok := signer.template.(*passkit.InMemoryPassTemplate)
	if !ok {
		t.Fatalf("expected in-memory template, got %T", signer.template)
	}

	files, err := memTemplate.GetAllFiles()
	if err != nil {
		t.Fatalf("get template files: %v", err)
	}

	expectedAssets := map[string]struct{}{
		"footer@3x.png": {},
		"icon@3x.png":   {},
		"logo@3x.png":   {},
	}

	if len(files) != len(expectedAssets) {
		t.Fatalf("unexpected asset count: %d", len(files))
	}

	for asset := range expectedAssets {
		if _, ok := files[asset]; !ok {
			t.Fatalf("missing embedded asset %q", asset)
		}
	}
	if _, ok := files["pass.json"]; ok {
		t.Fatalf("pass definition should not be part of template payload")
	}
}

func TestEmbeddedCreatorUsesFirstAndLastName(t *testing.T) {
	logger.Init()
	t.Helper()

	signer := &capturingSigner{}
	signingLoader := func(_, _, _ string) (*passkit.SigningInformation, error) {
		return &passkit.SigningInformation{}, nil
	}

	creator := NewEmbeddedApplePassCreator(
		AppleConfig{
			PassTypeIdentifier:         "pass.com.hakuna.integration",
			TeamIdentifier:             "TEAMHAKUNA",
			OrganizationName:           "Hakuna Wallet",
			Description:                "Hakuna Wallet Ticket",
			LogoText:                   "Hakuna",
			SigningCertificatePath:     "/tmp/cert.p12",
			SigningCertificatePassword: "integration-password",
			AppleRootCertificatePath:   "/tmp/root.cer",
		},
	)
	creator.Signer = signer
	creator.SigningInfoLoader = signingLoader

	ticket := tickets.TTIssuedTicket{
		ID:        "tt_embed_002",
		Barcode:   "EMBED-QR-002",
		FirstName: "Timon",
		LastName:  "Meerkat",
	}

	if _, err := creator.Create(context.Background(), ticket); err != nil {
		t.Fatalf("generate embedded pass: %v", err)
	}

	pass := signer.pass
	if pass == nil {
		t.Fatalf("signer did not capture pass")
	}

	nameFound := false
	for _, field := range pass.BoardingPass.SecondaryFields {
		if strings.EqualFold(field.Label, "passenger") {
			nameFound = true
			if field.Value != "Timon Meerkat" {
				t.Fatalf("expected passenger name Timon Meerkat, got %v", field.Value)
			}
		}
	}

	if !nameFound {
		t.Fatalf("expected passenger field to be updated")
	}
}

func TestEmbeddedCreatorRequiresTicketHolderName(t *testing.T) {
	logger.Init()
	t.Helper()

	creator := NewEmbeddedApplePassCreator(
		AppleConfig{
			PassTypeIdentifier:         "pass.com.hakuna.integration",
			TeamIdentifier:             "TEAMHAKUNA",
			OrganizationName:           "Hakuna Wallet",
			Description:                "Hakuna Wallet Ticket",
			LogoText:                   "Hakuna",
			SigningCertificatePath:     "/tmp/cert.p12",
			SigningCertificatePassword: "integration-password",
			AppleRootCertificatePath:   "/tmp/root.cer",
		},
	)

	ticket := tickets.TTIssuedTicket{
		ID:      "tt_embed_003",
		Barcode: "EMBED-QR-003",
	}

	_, err := creator.Create(context.Background(), ticket)
	if err == nil {
		t.Fatalf("expected error when name is missing")
	}
	if !strings.Contains(err.Error(), "ticket holder name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmbeddedCreatorRequiresBarcode(t *testing.T) {
	logger.Init()
	t.Helper()

	creator := NewEmbeddedApplePassCreator(
		AppleConfig{
			PassTypeIdentifier:         "pass.com.hakuna.integration",
			TeamIdentifier:             "TEAMHAKUNA",
			OrganizationName:           "Hakuna Wallet",
			Description:                "Hakuna Wallet Ticket",
			LogoText:                   "Hakuna",
			SigningCertificatePath:     "/tmp/cert.p12",
			SigningCertificatePassword: "integration-password",
			AppleRootCertificatePath:   "/tmp/root.cer",
		},
	)

	ticket := tickets.TTIssuedTicket{
		ID:       "tt_embed_004",
		FullName: "Pumba Warthog",
	}

	_, err := creator.Create(context.Background(), ticket)
	if err == nil {
		t.Fatalf("expected error when barcode is missing")
	}
	if !strings.Contains(err.Error(), "ticket barcode is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
