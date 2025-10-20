package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPassRepositories(t *testing.T) {
	ctx := context.Background()
	conn := setupTestDatabase(t, ctx)

	producedAt := time.Date(2025, 10, 20, 12, 0, 0, 0, time.UTC)
	err := MarkPassProduced(ctx, conn, "apple_wallet", "tt_ticket_1", "buyer@example.com", producedAt)
	require.NoError(t, err)

	err = MarkPassProduced(ctx, conn, "apple_wallet", "tt_ticket_2", "carol@example.com", producedAt)
	require.NoError(t, err)
	err = mutatePass(ctx, conn, "tt_ticket_2", "apple_wallet", func(pass *TicketPass) {
		pass.Status = string(Sent)
	})
	require.NoError(t, err)

	err = MarkPassProduced(ctx, conn, "apple_wallet", "tt_ticket_3", "pending@example.com", producedAt)
	require.NoError(t, err)
	err = mutatePass(ctx, conn, "tt_ticket_3", "apple_wallet", func(pass *TicketPass) {
		pass.Status = string(Pending)
		pass.ProducedAt = nil
	})
	require.NoError(t, err)

	err = MarkPassProduced(ctx, conn, "google_wallet", "tt_ticket_1", "buyer@example.com", producedAt)
	require.NoError(t, err)

	records, err := ListProducedPasses(ctx, conn, "apple_wallet")
	require.NoError(t, err)

	require.Len(t, records, 2)

	record1, ok := records["tt_ticket_1"]
	require.True(t, ok, "ticket 1 should be present")
	require.Equal(t, "buyer@example.com", record1.PurchaserEmail)
	require.Equal(t, string(Produced), record1.Status)
	require.True(t, producedAt.Equal(*record1.ProducedAt))

	record2, ok := records["tt_ticket_2"]
	require.True(t, ok, "ticket 2 should be present")
	require.Equal(t, string(Sent), record2.Status)
	require.Nil(t, record2.ErrorMessage)
	require.NotNil(t, record2.ProducedAt)

	_, exists := records["tt_ticket_3"]
	require.False(t, exists, "pending pass should not appear")

	googleRecords, err := ListProducedPasses(ctx, conn, "google_wallet")
	require.NoError(t, err)
	require.Len(t, googleRecords, 1)
	googleRecord, ok := googleRecords["tt_ticket_1"]
	require.True(t, ok, "google pass should exist")
	require.Equal(t, string(Produced), googleRecord.Status)
}

func mutatePass(ctx context.Context, conn *gorm.DB, ticketTailorID string, channel string, mutate func(*TicketPass)) error {
	var pass TicketPass
	err := conn.WithContext(ctx).
		Table("ticket_passes").
		Joins("JOIN tickets ON tickets.id = ticket_passes.ticket_id").
		Where("tickets.ticket_tailor_id = ? AND ticket_passes.channel = ?", ticketTailorID, channel).
		Select("ticket_passes.*").
		First(&pass).Error
	if err != nil {
		return err
	}

	mutate(&pass)
	return conn.WithContext(ctx).Save(&pass).Error
}
