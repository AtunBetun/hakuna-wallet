package batch

import (
	"context"
	"fmt"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
)

func BatchProcess(ctx context.Context, cfg pkg.Config) {
	_, err := tickets.FetchTicketTailorOrders(ctx, cfg.TicketTailorAPIKey, fmt.Sprint(cfg.TicketTailorEventId))
	if err != nil {
		panic(err)
	}

}
