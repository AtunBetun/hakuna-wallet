# AI Agent Prompt Template for Ticket Persistence Integration

## Role
You are a senior Golang engineer tasked with introducing durable ticket persistence using PostgreSQL (Neon) and GORM, while honouring this codebase’s functional style.

## Goals
[ ] Model a minimal schema that records Ticket Tailor ticket IDs plus wallet-pass production metadata per channel (Apple & Google)  
[ ] Bootstrap a Neon-hosted PostgreSQL instance and surface connection details via `pkg.Config`  
[ ] Implement GORM repositories that remain pure by injecting dependencies and returning errors  
[ ] Update the batch job pipeline to:
    1. Fetch all tickets from Ticket Tailor
    2. Load already-produced ticket IDs from Postgres
    3. Generate and send wallet passes only for new tickets, then persist their per-channel production state  
[ ] Provide migration strategy and developer setup docs (SQL or `gormigrate`) without mutating shared state implicitly

## Constraints / Preferences
- Keep data access points thin, favour small repository interfaces returning immutable DTOs.  
- Assume Neon uses the standard PostgreSQL wire protocol; no vendor-specific features needed.  
- Transactions should live at call sites; repositories accept `context.Context` and *gorm.DB to preserve composability.  
- Avoid hidden globals—wire GORM through constructors and configuration structs.  
- Tests should prefer an in-memory `sqlmock` (or dockerised Postgres if necessary) to keep CI deterministic.

## Environment / Context
- Runtime is deployed on Fly.io; DATABASE_URL already lives in `fly.toml`; ensure the new schema works with that connection string.  
- `pkg.Config` is the single source of truth for env vars. Keep Neon connection strings raw (no hidden parsing) and expose a `Validate()` method that derives computed fields after checking inputs.  
- `cmd/batch` orchestrates the ticket run; add persistence without breaking existing Apple Wallet generation flow.  
- The `tickets` directory stores generated artifacts; database rows should reference the Ticket Tailor ID and produced timestamp, not the file path.

## Output Requirements
- SQL migration or GORM `AutoMigrate` snippet that creates the minimal tables:
  - `tickets` table with immutable ticket-level data.
  - `ticket_passes` table keyed by ticket + channel to track Apple/Google status separately.
- GORM models and repository interfaces under `pkg/db` or a dedicated persistence package, plus unit tests.  
- Batch workflow changes that combine remote API data with persisted state, returning slices for downstream processing.  
- README or `docs/` note that explains how to run migrations locally (Neon connection string, `go run` command, etc.).
- Makefile targets (`migrate-up`, `migrate-down`) that wrap the `golang-migrate` CLI using the raw `DATABASE_URL`.

### Reference Schema
```sql
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_tailor_id TEXT NOT NULL UNIQUE,
    purchaser_email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ticket_passes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    channel TEXT NOT NULL, -- e.g. 'apple_wallet', 'google_wallet'
    status TEXT NOT NULL,  -- e.g. 'pending', 'produced', 'sent', 'failed'
    produced_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    error_message TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ticket_id, channel)
);
```

## Autonomy Instructions
- Before coding, confirm assumptions about Ticket Tailor identifiers (string vs int) and email fields.  
- Surface follow-up tasks (e.g., background retries, cleanup jobs) in `tasks.md` once gaps are discovered.  
- Keep SQL explicit—no magic migrations in production; version control all schema changes.  
- Ensure logging stays via `zap`; record meaningful fields like ticket IDs and status transitions.  
- Prefer idempotent operations so rerunning the batch doesn’t resend already-produced passes.
