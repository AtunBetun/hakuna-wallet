package db

import (
	"time"

	"gorm.io/datatypes"
)

// Ticket represents the ticket_tailor record persisted in the database.
type Ticket struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TicketTailorID string    `gorm:"column:ticket_tailor_id;type:text;uniqueIndex;not null"`
	PurchaserEmail string    `gorm:"column:purchaser_email;type:text;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

// TableName overrides the default table name.
func (Ticket) TableName() string {
	return "tickets"
}

// TicketPass models the per-channel wallet pass production metadata.
type TicketPass struct {
	ID           string            `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TicketID     string            `gorm:"column:ticket_id;type:uuid;not null;index:idx_ticket_passes_ticket_channel,priority:1"`
	Channel      string            `gorm:"column:channel;type:text;not null;index:idx_ticket_passes_ticket_channel,priority:2"`
	Status       string            `gorm:"column:status;type:text;not null;index:idx_ticket_passes_status"`
	ProducedAt   *time.Time        `gorm:"column:produced_at;type:timestamptz"`
	DeliveredAt  *time.Time        `gorm:"column:delivered_at;type:timestamptz"`
	ErrorMessage *string           `gorm:"column:error_message;type:text"`
	Metadata     datatypes.JSONMap `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt    time.Time         `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt    time.Time         `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

// TableName overrides the default table name.
func (TicketPass) TableName() string {
	return "ticket_passes"
}
