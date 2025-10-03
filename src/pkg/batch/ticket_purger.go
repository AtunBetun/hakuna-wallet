package batch

import (
	"context"
	"fmt"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"go.uber.org/zap"
)

func PurgeTickets(ctx context.Context, cfg pkg.Config) {
	ticketTailorConfig := tickets.TicketTailorConfig{
		ApiKey:  cfg.TicketTailorAPIKey,
		EventId: fmt.Sprint(cfg.TicketTailorEventId),
		BaseUrl: cfg.TicketTailorBaseUrl,
	}
	err := ticketTailorConfig.Validate()
	if err != nil {
		panic(err)
	}

	tick, err := tickets.FetchIssuedTickets(ctx, ticketTailorConfig)
	if err != nil {
		panic(err)
	}
	for _, v := range tick {
		if v.CheckedIn == "false" {
			logger.Logger.Debug("Unchecking ticket", zap.Any("ticketId", v.ID))
			tickets.CheckInTicket(ctx, ticketTailorConfig, v.ID, tickets.CheckOut)
		}
	}

}
