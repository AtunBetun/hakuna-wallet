-- Enable UUID generation for primary keys
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_tailor_id TEXT NOT NULL UNIQUE,
    purchaser_email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_ticket_tailor_id ON tickets (ticket_tailor_id);

CREATE TABLE IF NOT EXISTS ticket_passes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    status TEXT NOT NULL,
    produced_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    error_message TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ticket_id, channel)
);

CREATE INDEX IF NOT EXISTS idx_ticket_passes_ticket_channel ON ticket_passes (ticket_id, channel);
CREATE INDEX IF NOT EXISTS idx_ticket_passes_status ON ticket_passes (status);
