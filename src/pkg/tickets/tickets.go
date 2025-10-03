package tickets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"go.uber.org/zap"
)

// Simplified Ticket Tailor order structure (map fields to what you need)
type TTOrder struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	QR    string `json:"qr"`
}

// FetchTicketTailorOrders - placeholder: replace with the real TT API call
func FetchTicketTailorOrders(ctx context.Context, apiKey string, eventId string) ([]TTOrder, error) {
	if apiKey == "" || eventId == "" {
		return nil, fmt.Errorf("TICKETTAILOR_API_KEY or TT_EVENT_ID not set")
	}

	// Example endpoint - Ticket Tailor API docs: adapt as needed
	url := fmt.Sprintf("https://api.tickettailor.com/v1/issued_tickets?status=completed&event_id=%s", eventId)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte(apiKey))
	req.Header.Set("Authorization", "Basic "+encodedApiKey)
	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger.Logger.Info("ticket tailor issued tickets", zap.Any("tickets", resp.Body))

	var payload struct {
		Orders []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			// For demo, assume api gives a barcode string
			Barcode string `json:"barcode"`
		} `json:"orders"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	logger.Logger.Info("parsed ticktets", zap.Any("tickets", payload))

	out := make([]TTOrder, 0, len(payload.Orders))
	for _, o := range payload.Orders {
		out = append(out, TTOrder{
			ID:    o.ID,
			Email: o.Email,
			Name:  o.Name,
			QR:    o.Barcode,
		})
	}
	return out, nil
}
