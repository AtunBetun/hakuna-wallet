package tickets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Logger = zap.NewNop()
	os.Exit(m.Run())
}

func TestFetchIssuedTicketsSuccess(t *testing.T) {
	var capturedAuth string
	var capturedQuery url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/issued_tickets" {
			t.Fatalf("expected path /issued_tickets, got %s", r.URL.Path)
		}

		capturedAuth = r.Header.Get("Authorization")
		capturedQuery = r.URL.Query()

		resp := TTResponse{
			Data: []TTIssuedTicket{
				{ID: "ticket-1"},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	config := TicketTailorConfig{
		ApiKey:  "secret-key",
		EventId: "event-123",
		BaseUrl: server.URL,
	}

	tickets, err := FetchIssuedTickets(context.Background(), config, "issued", "cursor-42")
	if err != nil {
		t.Fatalf("FetchIssuedTickets returned error: %v", err)
	}

	if len(tickets) != 1 || tickets[0].ID != "ticket-1" {
		t.Fatalf("unexpected tickets returned: %+v", tickets)
	}

	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(config.ApiKey))
	if capturedAuth != expectedAuth {
		t.Fatalf("unexpected Authorization header. want %q, got %q", expectedAuth, capturedAuth)
	}

	if capturedQuery.Get("event_id") != config.EventId {
		t.Fatalf("event_id query mismatch. want %q, got %q", config.EventId, capturedQuery.Get("event_id"))
	}
	if capturedQuery.Get("status") != "issued" {
		t.Fatalf("status query mismatch. want %q, got %q", "issued", capturedQuery.Get("status"))
	}
	if capturedQuery.Get("starting_after") != "cursor-42" {
		t.Fatalf("starting_after query mismatch. want %q, got %q", "cursor-42", capturedQuery.Get("starting_after"))
	}
}

func TestFetchIssuedTicketsInvalidConfig(t *testing.T) {
	_, err := FetchIssuedTickets(context.Background(), TicketTailorConfig{}, "issued", "")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestFetchAllIssuedTickets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("event_id") != "event-123" {
			t.Fatalf("unexpected event_id. want %q, got %q", "event-123", query.Get("event_id"))
		}

		var tickets []TTIssuedTicket
		switch query.Get("starting_after") {
		case "":
			tickets = []TTIssuedTicket{{ID: "1"}, {ID: "2"}}
		case "2":
			tickets = []TTIssuedTicket{{ID: "3"}}
		default:
			tickets = nil
		}

		resp := TTResponse{Data: tickets}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	config := TicketTailorConfig{
		ApiKey:  "secret-key",
		EventId: "event-123",
		BaseUrl: server.URL,
	}

	tickets, err := FetchAllIssuedTickets(context.Background(), config, "issued")
	if err != nil {
		t.Fatalf("FetchAllIssuedTickets returned error: %v", err)
	}

	if len(tickets) != 3 {
		t.Fatalf("expected 3 tickets, got %d", len(tickets))
	}
	if tickets[0].ID != "1" || tickets[1].ID != "2" || tickets[2].ID != "3" {
		t.Fatalf("unexpected ticket order: %+v", tickets)
	}
}

func TestCheckInTicket(t *testing.T) {
	tests := []struct {
		name         string
		action       CheckAction
		wantQuantity string
	}{
		{name: "check in", action: CheckIn, wantQuantity: "1"},
		{name: "check out", action: CheckOut, wantQuantity: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedAuth string
			var receivedContentType string
			var formValues url.Values

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/check_ins" {
					t.Fatalf("expected path /check_ins, got %s", r.URL.Path)
				}

				if err := r.ParseForm(); err != nil {
					t.Fatalf("failed to parse form: %v", err)
				}

				receivedAuth = r.Header.Get("Authorization")
				receivedContentType = r.Header.Get("Content-Type")
				formValues = r.Form

				resp := CheckInResponse{
					ID:             "response-1",
					IssuedTicketID: r.Form.Get("issued_ticket_id"),
					Quantity:       0,
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			config := TicketTailorConfig{
				ApiKey:  "secret-key",
				EventId: "event-123",
				BaseUrl: server.URL,
			}

			resp, err := CheckInTicket(context.Background(), config, "ticket-42", tt.action)
			if err != nil {
				t.Fatalf("CheckInTicket returned error: %v", err)
			}

			expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(config.ApiKey))
			if receivedAuth != expectedAuth {
				t.Fatalf("unexpected Authorization header. want %q, got %q", expectedAuth, receivedAuth)
			}

			if receivedContentType != "application/x-www-form-urlencoded" {
				t.Fatalf("unexpected Content-Type header. want %q, got %q", "application/x-www-form-urlencoded", receivedContentType)
			}

			if formValues.Get("issued_ticket_id") != "ticket-42" {
				t.Fatalf("unexpected issued_ticket_id. want %q, got %q", "ticket-42", formValues.Get("issued_ticket_id"))
			}

			if formValues.Get("quantity") != tt.wantQuantity {
				t.Fatalf("unexpected quantity. want %q, got %q", tt.wantQuantity, formValues.Get("quantity"))
			}

			if resp.ID != "response-1" || resp.IssuedTicketID != "ticket-42" {
				t.Fatalf("unexpected CheckInResponse: %+v", resp)
			}
		})
	}
}

func TestCheckInTicketInvalidConfig(t *testing.T) {
	_, err := CheckInTicket(context.Background(), TicketTailorConfig{}, "ticket-42", CheckIn)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
