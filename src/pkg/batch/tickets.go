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

	ticketTailorConfig, err := tickets.NewTicketTailorConfig(cfg)
	if err != nil {
		panic(err)
	}
	err = ticketTailorConfig.Validate()
	if err != nil {
		panic(err)
	}

	tick, err := tickets.FetchAllIssuedTickets(ctx, ticketTailorConfig, "issued")
	if err != nil {
		panic(err)
	}

	logger.Logger.Debug("Count", zap.Any("tick count", len(tick)))
	for i, v := range tick {
		var check tickets.CheckAction
		check = tickets.CheckIn
		logger.Logger.Debug(fmt.Sprintf("%s ticket, loop: %d", check, i), zap.Any("ticketId", v.ID))
		tickets.CheckInTicket(ctx, ticketTailorConfig, v.ID, check)
	}

}

func GenerateTickets(ctx context.Context, cfg pkg.Config) {
	logger.Logger.Info("Generating Apple Wallet Tickets")
	ticketGenerator, err := NewWalletTicketGenerator(cfg)
	if err != nil {
		panic(err)
	}
	_, err = ticketGenerator.GenerateTickets(ctx)
	if err != nil {
		panic(err)
	}
	logger.Logger.Info("Finished Apple Wallet Tickets")

}
