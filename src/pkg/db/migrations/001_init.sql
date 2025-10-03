CREATE TABLE IF NOT EXISTS tickets (
  id BIGSERIAL PRIMARY KEY,
  ticket_tailor_id TEXT UNIQUE NOT NULL,
  purchaser_email TEXT,
  purchaser_name TEXT,
  qr_string TEXT NOT NULL,
  pkpass_path TEXT,
  google_jwt TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  retry_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);

