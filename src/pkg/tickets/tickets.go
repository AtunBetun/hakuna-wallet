package tickets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/atunbetun/hakuna-wallet/pkg/http_logs"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"go.uber.org/zap"
)

// TODO: this doesn't enforce compile time checks for values, which is sad

type TicketStatus string

const (
	Valid TicketStatus = "valid"
)

type CheckAction string

const (
	CheckIn  CheckAction = "checkIn"
	CheckOut CheckAction = "checkOut"
)

func FetchIssuedTickets(
	ctx context.Context,
	config TicketTailorConfig,
	status TicketStatus,
	startingAfter string,
) (
	[]TTIssuedTicket,
	error,
) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(config.BaseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "issued_tickets")
	q := url.Values{}
	q.Set("event_id", config.EventId)
	q.Set("status", string(status))
	if startingAfter != "" {
		q.Set("starting_after", startingAfter)
	}
	u.RawQuery = q.Encode()

	logger.Logger.Debug("TT Url", zap.Any("url", u.String()))

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte(config.ApiKey))
	req.Header.Set("Authorization", "Basic "+encodedApiKey)

	// TODO: this should probably be injected
	client := http_logs.NewLoggingClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ttResp TTResponse
	err = json.Unmarshal(body, &ttResp)
	if err != nil {
		return nil, err
	}
	// logger.Logger.Debug("TT Issued Tickets", zap.Any("data", ttResp.Data))

	return ttResp.Data, nil
}

func FetchAllIssuedTickets(
	ctx context.Context,
	config TicketTailorConfig,
	status TicketStatus,
) (
	[]TTIssuedTicket,
	error,
) {
	var allTickets []TTIssuedTicket
	var startingAfter string

	for {
		tickets, err := FetchIssuedTickets(ctx, config, status, startingAfter)
		if err != nil {
			return nil, err
		}

		if len(tickets) == 0 {
			break
		}

		allTickets = append(allTickets, tickets...)
		startingAfter = tickets[len(tickets)-1].ID
	}

	return allTickets, nil
}

func CheckInTicket(
	ctx context.Context,
	config TicketTailorConfig,
	ticketId string,
	checkAction CheckAction,
) (
	CheckInResponse,
	error,
) {
	if err := config.Validate(); err != nil {
		return CheckInResponse{}, err
	}

	u, err := url.Parse(config.BaseUrl)
	if err != nil {
		return CheckInResponse{}, err
	}

	var quantity int
	switch checkAction {
	case CheckIn:
		quantity = 1
	case CheckOut:
		quantity = -1
	}

	payload := strings.NewReader(fmt.Sprintf("issued_ticket_id=%s&quantity=%d", ticketId, quantity))
	u.Path = path.Join(u.Path, "check_ins")

	logger.Logger.Debug("TT Url", zap.Any("url", u.String()))

	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), payload)
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte(config.ApiKey))
	req.Header.Set("Authorization", "Basic "+encodedApiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// TODO: this should probably be injected
	client := http_logs.NewLoggingClient()
	resp, err := client.Do(req)
	if err != nil {
		return CheckInResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var chResponse CheckInResponse
	err = json.Unmarshal(body, &chResponse)
	if err != nil {
		panic(err)
	}
	logger.Logger.Debug(fmt.Sprintf("TT %s ticket", checkAction), zap.Any("ticketId", ticketId), zap.Any("action", checkAction))

	return chResponse, nil
}
