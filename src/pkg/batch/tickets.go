package batch

import (
	"context"
	"fmt"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/atunbetun/hakuna-wallet/pkg/db"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"github.com/atunbetun/hakuna-wallet/pkg/tickets"
	"go.uber.org/zap"
)

func PurgeTickets(ctx context.Context, cfg pkg.AppConfig) {
	ticketTailorConfig, err := tickets.NewTicketTailorConfig(cfg)
	if err != nil {
		panic(err)
	}
	err = ticketTailorConfig.Validate()
	if err != nil {
		panic(err)
	}

	tick, err := tickets.FetchAllIssuedTickets(ctx, ticketTailorConfig, "")
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

func GenerateTickets(ctx context.Context, cfg pkg.AppConfig) error {
	databaseCfg, err := db.FromAppConfig(cfg)
	if err != nil {
		return err
	}

	conn, err := db.Open(ctx, databaseCfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(conn); err != nil {
			panic(err)
		}
	}()

	ticketGenerator, err := NewWalletTicketSyncer(ctx, cfg, conn)
	if err != nil {
		return err
	}
	logger.Logger.Info("Syncing tickets")
	_, err = ticketGenerator.SyncTickets(ctx)
	if err != nil {
		return err
	}
	logger.Logger.Info("Finished Apple Wallet Tickets")
	return nil
}
