package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PassStatus string

const (
	Pending  PassStatus = "pending"
	Produced PassStatus = "produced"
	Sent     PassStatus = "sent"
	Failed   PassStatus = "failed"
)

type PassChannel string

const (
	AppleWalletChannel  PassChannel = "apple_wallet"
	GoogleWalletChannel PassChannel = "google_wallet"
)

// PassRecord captures the persisted state of a wallet pass for a given ticket.
type PassRecord struct {
	TicketTailorID string
	PurchaserEmail string
	Status         string
	ProducedAt     *time.Time
	DeliveredAt    *time.Time
	ErrorMessage   *string
}

// GetProducedPasses returns a map keyed by Ticket Tailor ID for passes that have been produced (or beyond) for a channel.
func GetProducedPasses(
	ctx context.Context,
	conn *gorm.DB,
	channel PassChannel,
) (map[string]PassRecord, error) {
	if conn == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	if channel == "" {
		return nil, fmt.Errorf("channel is required")
	}

	type ticketsRow struct {
		TicketTailorID string     `gorm:"column:ticket_tailor_id"`
		PurchaserEmail string     `gorm:"column:purchaser_email"`
		Status         string     `gorm:"column:status"`
		ProducedAt     *time.Time `gorm:"column:produced_at"`
		DeliveredAt    *time.Time `gorm:"column:delivered_at"`
		ErrorMessage   *string    `gorm:"column:error_message"`
	}

	var results []ticketsRow
	err := conn.WithContext(ctx).
		Table("ticket_passes").
		Select("tickets.ticket_tailor_id", "tickets.purchaser_email", "ticket_passes.status", "ticket_passes.produced_at", "ticket_passes.delivered_at", "ticket_passes.error_message").
		Joins("JOIN tickets ON tickets.id = ticket_passes.ticket_id").
		Where("ticket_passes.channel = ?", channel).
		Where("ticket_passes.status IN ?", []string{string(Produced), string(Sent)}).
		Find(&results).
		Error
	if err != nil {
		return nil, fmt.Errorf("listing produced passes: %w", err)
	}

	records := make(map[string]PassRecord, len(results))
	for _, r := range results {
		records[r.TicketTailorID] = PassRecord{
			TicketTailorID: r.TicketTailorID,
			PurchaserEmail: r.PurchaserEmail,
			Status:         r.Status,
			ProducedAt:     r.ProducedAt,
			DeliveredAt:    r.DeliveredAt,
			ErrorMessage:   r.ErrorMessage,
		}
	}

	return records, nil
}

// SetPassProduced upserts the ticket and marks the channel-specific pass as produced.
func SetPassProduced(
	ctx context.Context,
	conn *gorm.DB,
	channel PassChannel,
	ticketTailorID string,
	email string,
	producedAt time.Time,
) error {
	if conn == nil {
		return fmt.Errorf("database connection is required")
	}
	if channel == "" {
		return fmt.Errorf("channel is required")
	}
	if ticketTailorID == "" {
		return fmt.Errorf("ticketTailorID is required")
	}
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if producedAt.IsZero() {
		return fmt.Errorf("producedAt must be set")
	}

	err := conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ticket Ticket
		err := tx.Clauses(
			clause.Locking{Strength: "UPDATE"}).
			Where("ticket_tailor_id = ?", ticketTailorID).
			First(&ticket).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			ticket = Ticket{
				TicketTailorID: ticketTailorID,
				PurchaserEmail: email,
			}
			if err := tx.Create(&ticket).Error; err != nil {
				return fmt.Errorf("creating ticket record: %w", err)
			}
		case err != nil:
			return fmt.Errorf("fetching ticket: %w", err)
		default:
			if email != "" && email != ticket.PurchaserEmail {
				ticket.PurchaserEmail = email
				if err := tx.Save(&ticket).Error; err != nil {
					return fmt.Errorf("updating ticket: %w", err)
				}
			}
		}

		var pass TicketPass
		err = tx.Where("ticket_id = ? AND channel = ?", ticket.ID, channel).First(&pass).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			pass = TicketPass{
				TicketID:     ticket.ID,
				Channel:      string(channel),
				Status:       string(Produced),
				ProducedAt:   &producedAt,
				ErrorMessage: nil,
			}
			if err := tx.Create(&pass).Error; err != nil {
				return fmt.Errorf("creating ticket pass: %w", err)
			}
		case err != nil:
			return fmt.Errorf("fetching ticket pass: %w", err)
		default:
			pass.Status = string(Produced)
			pass.ProducedAt = &producedAt
			pass.ErrorMessage = nil
			if err := tx.Save(&pass).Error; err != nil {
				return fmt.Errorf("updating ticket pass: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
