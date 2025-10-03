package tickets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"go.uber.org/zap"
)

func FetchTicketTailorOrders(ctx context.Context, config TicketTailorConfig) ([]TTIssuedTicket, error) {
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
