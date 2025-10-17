package apple_test

import (
	"context"
	"testing"

	"github.com/alvinbaena/passkit"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet/apple"
)

type capturingSigner struct {
	pass      *passkit.Pass
	template  passkit.PassTemplate
	signing   *passkit.SigningInformation
	payload   []byte
	callCount int
}

func (c *capturingSigner) CreateSignedAndZippedPassArchive(pass *passkit.Pass, template passkit.PassTemplate, info *passkit.SigningInformation) ([]byte, error) {
	c.pass = pass
	c.template = template
	c.signing = info
	c.callCount++
	c.payload = []byte("signed-pass-" + pass.SerialNumber)
	return c.payload, nil
}

func TestBuildSignedPass(t *testing.T) {
	logger.Init()
	t.Helper()

	signer := &capturingSigner{}
	signingLoader := func(_, _, _ string) (*passkit.SigningInformation, error) {
		return &passkit.SigningInformation{}, nil
	}

	gen := apple.NewApplePassCreator(
		apple.AppleConfig{
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
	gen.Signer = signer
	gen.SigningInfoLoader = signingLoader
	gen.QRSize = 120

	ticket := tickets.TTIssuedTicket{
		ID:          "tt_555",
		Barcode:     "BR-555",
		Description: "Hakuna Integration Evening",
		FullName:    "Kiara Hakuna",
	}

	artifact, err := gen.Create(context.Background(), ticket)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if artifact.Platform != "apple" {
		t.Fatalf("expected platform apple, got %s", artifact.Platform)
	}
	if artifact.FileName != "tt_555.pkpass" {
		t.Fatalf("expected filename tt_555.pkpass, got %s", artifact.FileName)
	}
	if artifact.ContentType != "application/vnd.apple.pkpass" {
		t.Fatalf("unexpected content type: %s", artifact.ContentType)
	}
	if string(artifact.Data) != "signed-pass-tt_555" {
		t.Fatalf("unexpected payload %s", string(artifact.Data))
	}

	if signer.callCount != 1 {
		t.Fatalf("expected signer invoked once, got %d", signer.callCount)
	}
	if signer.pass == nil {
		t.Fatalf("signer did not capture pass")
	}
	if signer.pass.SerialNumber != "tt_555" {
		t.Fatalf("unexpected serial number %s", signer.pass.SerialNumber)
	}
	if signer.pass.OrganizationName != "Hakuna Wallet" {
		t.Fatalf("unexpected organization name %s", signer.pass.OrganizationName)
	}
	if signer.pass.Description != "Hakuna Wallet Ticket" {
		t.Fatalf("unexpected pass description %s", signer.pass.Description)
	}
	if len(signer.pass.Barcodes) != 1 {
		t.Fatalf("expected a single barcode, got %d", len(signer.pass.Barcodes))
	}
	if signer.pass.Barcodes[0].Message != "BR-555" {
		t.Fatalf("barcode payload mismatch: %s", signer.pass.Barcodes[0].Message)
	}
	if signer.signing == nil {
		t.Fatalf("signing information not passed to signer")
	}

	memTemplate, ok := signer.template.(*passkit.InMemoryPassTemplate)
	if !ok {
		t.Fatalf("expected in-memory template, got %T", signer.template)
	}

	files, err := memTemplate.GetAllFiles()
	if err != nil {
		t.Fatalf("get files: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("expected QR asset to be attached")
	}
}
