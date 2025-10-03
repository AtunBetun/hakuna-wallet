package tickets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Simplified Ticket Tailor order structure (map fields to what you need)
type TTOrder struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	QR    string `json:"qr"`
}

// FetchTicketTailorOrders - placeholder: replace with the real TT API call
func FetchTicketTailorOrders(ctx context.Context, since time.Time, apiKey string, eventId string) ([]TTOrder, error) {
	if apiKey == "" || eventId == "" {
		return nil, fmt.Errorf("TICKETTAILOR_API_KEY or TT_EVENT_ID not set")
	}

	// Example endpoint - Ticket Tailor API docs: adapt as needed
	url := fmt.Sprintf("https://api.tickettailor.com/v1/events/%s/issued_tickets?updated_after=%s&status=completed", eventId, since.Format(time.RFC3339))
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
