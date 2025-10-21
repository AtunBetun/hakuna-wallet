# Tasks

## Ticket lifecycle
- [ ] Stand up Neon Postgres instance dedicated to ticket issuance metadata and document connection string in `.env`
- [x] Add GORM dependency and bootstrap `gorm.DB` wiring through `pkg.Config` and `cmd/batch` constructors; keep Neon connection string raw in config and validate via `Config.Validate()`
- [x] Define migrations for `tickets` (immutable ticket data) and `ticket_passes` (channel-specific status) tables per `prompt.md`
- [x] Introduce `golang-migrate/migrate` CLI into workflow; add Makefile task to run migrations locally and in CI
- [x] Author initial `migrations/0001_init.sql` using `golang-migrate` format covering `tickets` and `ticket_passes`
- [x] Implement repository functions:
  - `ListProducedPasses(ctx context.Context, db *gorm.DB, channel string) (map[string]PassRecord, error)`
  - `MarkPassProduced(ctx context.Context, db *gorm.DB, channel string, ticketID string, email string, producedAt time.Time) error`
- [ ] Update batch pipeline to: fetch Ticket Tailor tickets, diff against repository data per channel (Apple, Google), produce passes, persist channel-specific metadata, and wrap DB writes in transactions
- [ ] Write diffing logic unit tests for batch pipeline
- [x] Write repository tests using testcontainers-backed Postgres database

## Ticket assets
- [x] Integrate the designer-provided ticket template into the generator
https://developer.apple.com/wallet/add-to-apple-wallet-guidelines/?utm_source=chatgpt.com

## Current embedded Apple pass plan
- [x] Review existing Apple wallet creator and ticket struct for required fields
- [x] Embed designer pass bundle within the codebase
- [x] Implement new Apple creator that reuses embedded assets and swaps QR/name
- [x] Format code and run `go test ./...`

## Email delivery
- [ ] Build email workflow to send the generated ticket

## Deployment
- [ ] Prepare Fly.io configuration (Dockerfile, fly.toml, secrets)
- [ ] Automate github deploy pipeline to Fly.io
- [ ] Set up semantic versioning workflow for releases
- [x] Gate Fly deploy workflow on successful CI run
- [x] Materialize certificate secrets into runtime files at startup
- [ ] Figure out the best cron setup

## Volume
- [ ] Figure out volume replication or go to s3
