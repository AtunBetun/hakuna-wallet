package apple

import (
	"context"

	"github.com/alvinbaena/passkit"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"github.com/atunbetun/hakuna-wallet/pkg/wallet"
)

// Signer defines the contract required to sign a pass archive.
type Signer interface {
	CreateSignedAndZippedPassArchive(pass *passkit.Pass, template passkit.PassTemplate, info *passkit.SigningInformation) ([]byte, error)
}

// SigningInfoLoader loads signing information from disk or another source.
type SigningInfoLoader func(certPath string, password string, rootPath string) (*passkit.SigningInformation, error)

// Creator exposes the minimum API needed by the rest of the system.
type Creator interface {
	Create(ctx context.Context, ticket tickets.TTIssuedTicket) (wallet.Artifact, error)
}

// AppleConfig collects the parameters required to render and sign an Apple Wallet pass.
type AppleConfig struct {
	PassTypeIdentifier         string `validate:"required"`
	TeamIdentifier             string `validate:"required"`
	OrganizationName           string
	Description                string
	LogoText                   string
	SigningCertificatePath     string `validate:"required"`
	SigningCertificatePassword string `validate:"required"`
	AppleRootCertificatePath   string `validate:"required"`
}
