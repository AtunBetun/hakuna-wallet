
package google_wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/walletobjects/v1"
)

const (
	classSuffix = "class"
	objectSuffix = "object"
)

// GoogleWalletService handles the creation of Google Wallet passes.

type GoogleWalletService struct {
	config *pkg.Config
	walletService *walletobjects.Service
}

// NewGoogleWalletService creates a new GoogleWalletService.

func NewGoogleWalletService(config *pkg.config) (*GoogleWalletService, error) {
	ctx := context.Background()

	// Read the service account key file
	key, err := os.ReadFile(config.GoogleServiceAccountJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key file: %w", err)
	}

	// Create new credentials
	creds, err := google.CredentialsFromJSON(ctx, key, "https://www.googleapis.com/auth/wallet_object.issuer")
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	// Create a new wallet service
	walletService, err := walletobjects.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet service: %w", err)
	}

	return &GoogleWalletService{
		config: config,
		walletService: walletService,
	}, nil
}

// CreateEventTicketClass creates a new event ticket class.

func (s *GoogleWalletService) CreateEventTicketClass(classId string) (string, error) {
	// Check if the class already exists
	_, err := s.walletService.Eventticketclass.Get(classId).Do()
	if err == nil {
		return fmt.Sprintf("Class %s already exists", classId), nil
	}

	// Create a new class
	newClass := &walletobjects.EventTicketClass{
		Id: classId,
		IssuerName: s.config.GoogleIssuerEmail,
		ReviewStatus: "under_review",
		EventName: &walletobjects.LocalizedString{
			DefaultValue: &walletobjects.TranslatedString{
				Language: "en-US",
				Value: "Hakuna Matata Festival",
			},
		},
	}

	_, err = s.walletService.Eventticketclass.Insert(newClass).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create class: %w", err)
	}

	return fmt.Sprintf("Class %s created successfully", classId), nil
}

// CreateEventTicketObject creates a new event ticket object.

func (s *GoogleWalletService) CreateEventTicketObject(classId, objectId string) (string, error) {
	// Check if the object already exists
	_, err := s.walletService.Eventticketobject.Get(objectId).Do()
	if err == nil {
		return fmt.Sprintf("Object %s already exists", objectId), nil
	}

	// Create a new object
	newObject := &walletobjects.EventTicketObject{
		Id: objectId,
		ClassId: classId,
		State: "active",
		HeroImage: &walletobjects.Image{
			SourceUri: &walletobjects.ImageUri{
				Uri: "https://example.com/hero.jpg",
			},
		},
		TextModulesData: []*walletobjects.TextModuleData{
			{
				Header: "Event Details",
				Body: "Hakuna Matata Festival",
			},
		},
		LinksModuleData: &walletobjects.LinksModuleData{
			Uris: []*walletobjects.Uri{
				{
					Uri: "https://example.com/tickets",
					Description: "View Tickets",
				},
			},
		},
		ImageModulesData: []*walletobjects.ImageModuleData{
			{
				MainImage: &walletobjects.Image{
					SourceUri: &walletobjects.ImageUri{
						Uri: "https://example.com/main.jpg",
					},
				},
			},
		},
		Barcode: &walletobjects.Barcode{
			Type: "qrCode",
			Value: "https://tickettailor.com/qr-code",
		},
		Locations: []*walletobjects.LatLongPoint{
			{
				Latitude: 37.424015499999996,
				Longitude: -122.09259560000001,
			},
		},
	}

	_, err = s.walletService.Eventticketobject.Insert(newObject).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create object: %w", err)
	}

	return fmt.Sprintf("Object %s created successfully", objectId), nil
}
