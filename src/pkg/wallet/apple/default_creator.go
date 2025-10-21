package apple

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image/png"

	"github.com/alvinbaena/passkit"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const (
	defaultQRSize     = 200
	qrMessageEncoding = "iso-8859-1"
)

//go:embed icon.png
var iconPNG []byte

// defaultApplePassCreator prepares Apple Wallet passes and signs them.
type defaultApplePassCreator struct {
	Config            AppleConfig `validate:"required"`
	Signer            Signer
	SigningInfoLoader SigningInfoLoader
	QRSize            int
}

// NewApplePassCreator returns a pass creator configured with the provided options.
func NewApplePassCreator(cfg AppleConfig) *defaultApplePassCreator {
	return &defaultApplePassCreator{
		Config: cfg,
	}
}

// passDraft keeps the prepared objects together while the pass is assembled.
type passDraft struct {
	Pass *passkit.Pass
	QR   []byte
}

// TODO: ehhh, not great
var validate = validator.New(validator.WithRequiredStructEnabled())

// Create builds, signs, and packages a pass for the provided ticket.
func (c *defaultApplePassCreator) Create(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return wallet.Artifact{}, err
	}
	logger.Logger.Debug(
		"creating apple wallet pass",
		zap.String("ticket_id", ticket.ID),
	)
	if err := validate.Struct(c.Config); err != nil {
		return wallet.Artifact{}, err
	}

	logger.Logger.Debug(
		"preparing pass draft",
		zap.String("ticket_id", ticket.ID),
		zap.Int("qr_size", c.qrSize()),
	)
	draft, err := preparePassDraft(c.Config, ticket, c.qrSize())
	if err != nil {
		return wallet.Artifact{}, err
	}

	logger.Logger.Debug(
		"loading signing information",
		zap.String("signing_certificate_path", c.Config.SigningCertificatePath),
	)
	signInfo, err := c.loadSigningInfo()
	if err != nil {
		return wallet.Artifact{}, fmt.Errorf("loading signing information: %w", err)
	}

	template := buildTemplate(draft.QR)
	logger.Logger.Debug(
		"signing ticket",
		zap.Any("ticket_id", ticket.ID),
	)
	payload, err := c.signer().CreateSignedAndZippedPassArchive(draft.Pass, template, signInfo)
	if err != nil {
		return wallet.Artifact{}, fmt.Errorf("signing pass: %w", err)
	}

	logger.Logger.Debug(
		"signed apple wallet pass",
		zap.String("ticket_id", ticket.ID),
	)

	return wallet.Artifact{
		Platform:    "apple",
		FileName:    fmt.Sprintf("%s.pkpass", ticket.ID),
		ContentType: "application/vnd.apple.pkpass",
		Data:        payload,
	}, nil
}

func (c *defaultApplePassCreator) qrSize() int {
	if c.QRSize > 0 {
		return c.QRSize
	}
	return defaultQRSize
}

func (c *defaultApplePassCreator) signer() Signer {
	if c.Signer != nil {
		return c.Signer
	}
	return passkit.NewMemoryBasedSigner()
}

func (c *defaultApplePassCreator) loadSigningInfo() (*passkit.SigningInformation, error) {
	loader := c.SigningInfoLoader
	if loader == nil {
		loader = passkit.LoadSigningInformationFromFiles
	}
	return loader(c.Config.SigningCertificatePath, c.Config.SigningCertificatePassword, c.Config.AppleRootCertificatePath)
}

func preparePassDraft(cfg AppleConfig, ticket tickets.TTIssuedTicket, qrSize int) (passDraft, error) {
	if ticket.Barcode == "" {
		return passDraft{}, fmt.Errorf("ticket barcode is required")
	}

	logger.Logger.Debug(
		"generating pass qr code",
		zap.String("ticket_id", ticket.ID),
		zap.Int("qr_size", qrSize),
	)
	qrBytes, err := generateQR(ticket.Barcode, qrSize)
	if err != nil {
		return passDraft{}, err
	}

	logger.Logger.Debug(
		"building apple wallet pass contents",
		zap.String("ticket_id", ticket.ID),
	)
	pass := buildPass(cfg, ticket)

	return passDraft{
		Pass: pass,
		QR:   qrBytes,
	}, nil
}

func generateQR(payload string, size int) ([]byte, error) {
	code, err := qr.Encode(payload, qr.M, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("encoding QR code: %w", err)
	}

	if size <= 0 {
		size = defaultQRSize
	}

	code, err = barcode.Scale(code, size, size)
	if err != nil {
		return nil, fmt.Errorf("scaling QR code: %w", err)
	}

	logger.Logger.Debug(
		"encoded qr to png",
		zap.Int("size", size),
	)

	var buf bytes.Buffer
	if err := png.Encode(&buf, code); err != nil {
		return nil, fmt.Errorf("encoding PNG: %w", err)
	}

	return buf.Bytes(), nil
}

func buildPass(cfg AppleConfig, ticket tickets.TTIssuedTicket) *passkit.Pass {
	logger.Logger.Debug(
		"assembling pass",
		zap.String("ticket_id", ticket.ID),
	)
	pass := &passkit.Pass{
		FormatVersion:      1,
		PassTypeIdentifier: cfg.PassTypeIdentifier,
		SerialNumber:       ticket.ID,
		TeamIdentifier:     cfg.TeamIdentifier,
		OrganizationName:   cfg.OrganizationName,
		Description:        cfg.Description,
		LogoText:           cfg.LogoText,
	}

	pass.Barcodes = append(pass.Barcodes, passkit.Barcode{
		Format:          passkit.BarcodeFormatQR,
		Message:         ticket.Barcode,
		MessageEncoding: qrMessageEncoding,
	})

	eventTicket := passkit.NewEventTicket()
	eventTicket.AddPrimaryFields(passkit.Field{
		Key:   "event",
		Label: "Event",
		Value: ticket.Description,
	})
	eventTicket.AddSecondaryFields(passkit.Field{
		Key:   "name",
		Label: "Name",
		Value: ticket.FullName,
	})
	pass.EventTicket = eventTicket

	return pass
}

func buildTemplate(qrBytes []byte) *passkit.InMemoryPassTemplate {
	template := passkit.NewInMemoryPassTemplate()
	template.AddFileBytes(passkit.BundleIcon, iconPNG)
	template.AddFileBytes(passkit.BundleThumbnail, qrBytes)
	logger.Logger.Debug("enriched pass template with icon and qr thumbnail")
	return template
}
