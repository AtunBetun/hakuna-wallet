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

	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"go.uber.org/zap"
)

// TODO: need to paginate and fetch all
func FetchIssuedTickets(ctx context.Context, config TicketTailorConfig) ([]TTIssuedTicket, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(config.BaseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "issued_tickets")
	// u.RawQuery = url.Values{"event_id": {config.EventId}}.Encode()

	logger.Logger.Debug("TT Url", zap.Any("url", u.String()))

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte(config.ApiKey))
	req.Header.Set("Authorization", "Basic "+encodedApiKey)

	resp, err := http.DefaultClient.Do(req)
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
	logger.Logger.Debug("TT Issued Tickets", zap.Any("data", ttResp.Data))

	return ttResp.Data, nil
}

type CheckAction string

const (
	CheckIn  CheckAction = "checkIn"
	CheckOut CheckAction = "checkOut"
)

func CheckInTicket(ctx context.Context, config TicketTailorConfig, ticketId string, checkAction CheckAction) (CheckInResponse, error) {
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckInResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CheckInResponse{}, err
	}

	var chResponse CheckInResponse
	err = json.Unmarshal(body, &chResponse)
	if err != nil {
		return CheckInResponse{}, err
	}
	logger.Logger.Debug(fmt.Sprintf("TT %s ticket", checkAction), zap.Any("ticketId", ticketId), zap.Any("action", checkAction))

	return chResponse, nil
}
