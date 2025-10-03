package batch

import (
	"context"
	"fmt"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
)

func BatchProcess(ctx context.Context, cfg pkg.Config) {
	ticketTailorConfig := tickets.TicketTailorConfig{
		ApiKey:  cfg.TicketTailorAPIKey,
		EventId: fmt.Sprint(cfg.TicketTailorEventId),
		BaseUrl: cfg.TicketTailorBaseUrl,
	}
	err := ticketTailorConfig.Validate()
	if err != nil {
		panic(err)
	}

	_, err = tickets.FetchTicketTailorOrders(ctx, ticketTailorConfig)
	if err != nil {
		panic(err)
	}

}
