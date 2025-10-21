package apple

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/alvinbaena/passkit"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
	"go.uber.org/zap"
)

//go:embed passbundle/*
var embeddedPassFiles embed.FS

const (
	embeddedPassDir       = "passbundle"
	passDefinitionFile    = "pass.json"
	manifestFileName      = "manifest.json"
	signatureFileName     = "signature"
	attendeeFieldFallback = "Passenger"
)

var skippedTemplateFiles = map[string]struct{}{
	passDefinitionFile: {},
	manifestFileName:   {},
	signatureFileName:  {},
}

// embeddedApplePassCreator produces Apple Wallet passes from an embedded designer bundle.
type embeddedApplePassCreator struct {
	Config            AppleConfig `validate:"required"`
	Signer            Signer
	SigningInfoLoader SigningInfoLoader
}

// NewEmbeddedApplePassCreator returns a creator that relies on the embedded pass assets.
func NewEmbeddedApplePassCreator(cfg AppleConfig) *embeddedApplePassCreator {
	return &embeddedApplePassCreator{Config: cfg}
}

// Create builds, signs, and packages a pass using the embedded pass template.
func (c *embeddedApplePassCreator) Create(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return wallet.Artifact{}, err
	}

	logger.Logger.Debug(
		"creating embedded apple wallet pass",
		zap.String("ticket_id", ticket.ID),
	)

	if err := validate.Struct(c.Config); err != nil {
		return wallet.Artifact{}, err
	}

	pass, err := c.buildPass(ticket)
	if err != nil {
		return wallet.Artifact{}, err
	}

	template, err := loadEmbeddedPassTemplate()
	if err != nil {
		return wallet.Artifact{}, fmt.Errorf("loading embedded template assets: %w", err)
	}

	logger.Logger.Debug("loading signing information for embedded pass")
	signInfo, err := c.loadSigningInfo()
	if err != nil {
		return wallet.Artifact{}, fmt.Errorf("loading signing information: %w", err)
	}

	payload, err := c.signer().CreateSignedAndZippedPassArchive(pass, template, signInfo)
	if err != nil {
		return wallet.Artifact{}, fmt.Errorf("signing pass: %w", err)
	}

	logger.Logger.Debug(
		"successfully created embedded apple pass",
		zap.String("ticket_id", ticket.ID),
	)

	return wallet.Artifact{
		Platform:    "apple",
		FileName:    fmt.Sprintf("%s.pkpass", ticket.ID),
		ContentType: "application/vnd.apple.pkpass",
		Data:        payload,
	}, nil
}

func (c *embeddedApplePassCreator) buildPass(ticket tickets.TTIssuedTicket) (*passkit.Pass, error) {
	logger.Logger.Debug(
		"preparing embedded pass definition",
		zap.String("ticket_id", ticket.ID),
	)

	raw, err := embeddedPassFiles.ReadFile(path.Join(embeddedPassDir, passDefinitionFile))
	if err != nil {
		return nil, fmt.Errorf("reading embedded pass definition: %w", err)
	}

	var pass passkit.Pass
	if err := json.Unmarshal(raw, &pass); err != nil {
		return nil, fmt.Errorf("decoding embedded pass definition: %w", err)
	}

	pass.PassTypeIdentifier = c.Config.PassTypeIdentifier
	pass.TeamIdentifier = c.Config.TeamIdentifier
	pass.OrganizationName = c.Config.OrganizationName
	pass.Description = c.Config.Description
	pass.LogoText = c.Config.LogoText
	pass.SerialNumber = ticket.ID

	// TODO: mutation
	if err := updatePassBarcode(&pass, ticket.Barcode); err != nil {
		return nil, err
	}

	// TODO: mutation
	if err := applyTicketHolderName(&pass, ticket); err != nil {
		return nil, err
	}

	return &pass, nil
}

func (c *embeddedApplePassCreator) signer() Signer {
	if c.Signer != nil {
		return c.Signer
	}
	return passkit.NewMemoryBasedSigner()
}

func (c *embeddedApplePassCreator) loadSigningInfo() (*passkit.SigningInformation, error) {
	loader := c.SigningInfoLoader
	if loader == nil {
		loader = passkit.LoadSigningInformationFromFiles
	}

	return loader(
		c.Config.SigningCertificatePath,
		c.Config.SigningCertificatePassword,
		c.Config.AppleRootCertificatePath,
	)
}

func loadEmbeddedPassTemplate() (*passkit.InMemoryPassTemplate, error) {
	template := passkit.NewInMemoryPassTemplate()

	entries, err := embeddedPassFiles.ReadDir(embeddedPassDir)
	if err != nil {
		return nil, fmt.Errorf("reading embedded pass directory: %w", err)
	}

	added := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if _, skip := skippedTemplateFiles[name]; skip {
			continue
		}

		data, err := embeddedPassFiles.ReadFile(path.Join(embeddedPassDir, name))
		if err != nil {
			return nil, fmt.Errorf("reading embedded asset %q: %w", name, err)
		}

		template.AddFileBytes(name, data)
		added++
	}

	logger.Logger.Debug(
		"assembled embedded template assets",
		zap.Int("asset_count", added),
	)

	return template, nil
}

// TODO: mutation
func updatePassBarcode(pass *passkit.Pass, raw string) error {
	code := strings.TrimSpace(raw)
	if code == "" {
		return fmt.Errorf("ticket barcode is required")
	}

	if len(pass.Barcodes) == 0 {
		pass.Barcodes = append(pass.Barcodes, passkit.Barcode{})
	}

	pass.Barcodes[0].Format = passkit.BarcodeFormatQR
	pass.Barcodes[0].Message = code
	pass.Barcodes[0].MessageEncoding = qrMessageEncoding
	pass.Barcodes[0].AltText = code

	return nil
}

// TODO: mutation
func applyTicketHolderName(pass *passkit.Pass, ticket tickets.TTIssuedTicket) error {
	name, err := resolveTicketHolderName(ticket)
	if err != nil {
		return err
	}

	if bp := pass.BoardingPass; bp != nil {
		if bp.GenericPass == nil {
			bp.GenericPass = passkit.NewGenericPass()
		}

		if updateFieldValue(bp.SecondaryFields, name) {
			return nil
		}

		bp.AddSecondaryFields(passkit.Field{
			Key:   "passenger",
			Label: attendeeFieldFallback,
			Value: name,
		})
		return nil
	}

	if et := pass.EventTicket; et != nil {
		if et.GenericPass == nil {
			et.GenericPass = passkit.NewGenericPass()
		}

		if updateFieldValue(et.SecondaryFields, name) {
			return nil
		}

		et.AddSecondaryFields(passkit.Field{
			Key:   "attendee",
			Label: attendeeFieldFallback,
			Value: name,
		})
		return nil
	}

	return fmt.Errorf("unable to apply ticket holder name; no compatible pass section present")
}

func resolveTicketHolderName(ticket tickets.TTIssuedTicket) (string, error) {
	name := strings.TrimSpace(ticket.FullName)
	if name != "" {
		return name, nil
	}

	first := strings.TrimSpace(ticket.FirstName)
	last := strings.TrimSpace(ticket.LastName)

	parts := make([]string, 0, 2)
	if first != "" {
		parts = append(parts, first)
	}
	if last != "" {
		parts = append(parts, last)
	}

	name = strings.TrimSpace(strings.Join(parts, " "))
	if name == "" {
		return "", fmt.Errorf("ticket holder name is required")
	}

	return name, nil
}

// TODO: mutation
func updateFieldValue(fields []passkit.Field, value string) bool {
	for i := range fields {
		label := strings.TrimSpace(fields[i].Label)
		key := strings.TrimSpace(fields[i].Key)

		if isPassengerField(label, key) {
			fields[i].Value = value
			return true
		}
	}
	return false
}

func isPassengerField(label, key string) bool {
	if label == "" && key == "" {
		return false
	}

	labelLower := strings.ToLower(label)
	keyLower := strings.ToLower(key)

	if labelLower == "passenger" ||
		strings.Contains(labelLower, "attendee") ||
		strings.Contains(labelLower, "name") {
		return true
	}

	if keyLower == "passenger" ||
		strings.Contains(keyLower, "passenger") ||
		strings.Contains(keyLower, "attendee") {
		return true
	}

	return false
}
