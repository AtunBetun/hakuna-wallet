package tickets

import (
	"fmt"

	"github.com/atunbetun/hakuna-wallet/pkg"
)

// Full struct mapping the API response
type TTResponse struct {
	Data []TTIssuedTicket `json:"data"`
}

type TTIssuedTicket struct {
	Object             string           `json:"object"`
	ID                 string           `json:"id"`
	AddOnID            *string          `json:"add_on_id"`
	Barcode            string           `json:"barcode"`
	BarcodeURL         string           `json:"barcode_url"`
	CheckedIn          string           `json:"checked_in"`
	CreatedAt          int64            `json:"created_at"`
	CustomQuestions    []any            `json:"custom_questions"`
	Description        string           `json:"description"`
	Email              string           `json:"email"`
	EventID            string           `json:"event_id"`
	EventSeriesID      string           `json:"event_series_id"`
	FirstName          string           `json:"first_name"`
	FullName           string           `json:"full_name"`
	GroupTicketBarcode *string          `json:"group_ticket_barcode"`
	LastName           string           `json:"last_name"`
	ListedCurrency     TTListedCurrency `json:"listed_currency"`
	ListedPrice        int              `json:"listed_price"`
	OrderID            string           `json:"order_id"`
	QRCodeURL          string           `json:"qr_code_url"`
	Reference          *string          `json:"reference"`
	Reservation        *string          `json:"reservation"`
	Source             string           `json:"source"`
	Status             string           `json:"status"`
	TicketTypeID       string           `json:"ticket_type_id"`
	UpdatedAt          int64            `json:"updated_at"`
	VoidedAt           *string          `json:"voided_at"`
}

type TTListedCurrency struct {
	BaseMultiplier int    `json:"base_multiplier"`
	Code           string `json:"code"`
}

type TicketTailorConfig struct {
	ApiKey  string
	EventId string
	BaseUrl string
}

func NewTicketTailorConfig(cfg pkg.Config) (TicketTailorConfig, error) {
	ticketCfg := TicketTailorConfig{
		ApiKey:  cfg.TicketTailorAPIKey,
		EventId: cfg.TicketTailorEventId,
		BaseUrl: cfg.TicketTailorBaseUrl,
	}

	if err := ticketCfg.Validate(); err != nil {
		return TicketTailorConfig{}, fmt.Errorf("invalid ticket tailor config: %w", err)
	}

	return ticketCfg, nil
}

func (c *TicketTailorConfig) Validate() error {
	if c.ApiKey == "" {
		return fmt.Errorf("ApiKey cannot be empty")
	}
	if c.EventId == "" {
		return fmt.Errorf("EventId cannot be empty")
	}
	if c.BaseUrl == "" {
		return fmt.Errorf("BaseUrl cannot be empty")
	}
	return nil
}

type CheckInResponse struct {
	Object         string `json:"object"`
	ID             string `json:"id"`
	CheckInAt      int64  `json:"check_in_at"`
	CreatedAt      int64  `json:"created_at"`
	EventID        string `json:"event_id"`
	EventSeriesID  string `json:"event_series_id"`
	IssuedTicketID string `json:"issued_ticket_id"`
	Quantity       int    `json:"quantity"`
}
