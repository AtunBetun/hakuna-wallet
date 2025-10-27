# Tasks

## Ticket lifecycle
- [x] Stand up Neon Postgres instance dedicated to ticket issuance metadata and document connection string in `.env`
- [x] Add GORM dependency and bootstrap `gorm.DB` wiring through `pkg.Config` and `cmd/batch` constructors; keep Neon connection string raw in config and validate via `Config.Validate()`
- [x] Define migrations for `tickets` (immutable ticket data) and `ticket_passes` (channel-specific status) tables per `prompt.md`
- [x] Introduce `golang-migrate/migrate` CLI into workflow; add Makefile task to run migrations locally and in CI
- [x] Author initial `migrations/0001_init.sql` using `golang-migrate` format covering `tickets` and `ticket_passes`
- [x] Implement repository functions:
  - `ListProducedPasses(ctx context.Context, db *gorm.DB, channel string) (map[string]PassRecord, error)`
  - `MarkPassProduced(ctx context.Context, db *gorm.DB, channel string, ticketID string, email string, producedAt time.Time) error`
- [x] Update batch pipeline to: fetch Ticket Tailor tickets, diff against repository data per channel (Apple, Google), produce passes, persist channel-specific metadata, and wrap DB writes in transactions
- [x] Write diffing logic unit tests for batch pipeline
- [x] Add wallet syncer integration tests for fresh and existing tickets using testcontainers + HTTP proxy
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
- [x] Build email workflow to send the generated ticket

## Deployment
- [ ] Set up semantic versioning workflow for releases
- [x] Prepare Fly.io configuration (Dockerfile, fly.toml, secrets)
- [x] Automate github deploy pipeline to Fly.io
- [x] Gate Fly deploy workflow on successful CI run
- [x] Materialize certificate secrets into runtime files at startup
- [x] Figure out the best cron setup

## Volume
- [x] Figure out volume replication or go to s3
- [x] Add better naming for how the tickets get saved

## Code TODO Backlog
- [ ] Refactor `pkg.Config.Validate` to eliminate mutation of receiver state (src/pkg/config.go:46)
- [ ] Enforce compile-time validation for ticket status values (src/pkg/tickets/tickets.go:19)
- [ ] Inject the HTTP client dependency in `FetchIssuedTickets` instead of constructing it inline (src/pkg/tickets/tickets.go:66)
- [ ] Inject the HTTP client dependency in `CheckInTicket` instead of constructing it inline (src/pkg/tickets/tickets.go:153)
- [ ] Remove the mutation warning when updating the embedded pass barcode assignment (src/pkg/wallet/apple/embedded_creator.go:119)
- [ ] Remove the mutation warning when applying the embedded pass ticket holder name (src/pkg/wallet/apple/embedded_creator.go:124)
- [ ] Refactor `updatePassBarcode` to avoid mutating the pass structure in place (src/pkg/wallet/apple/embedded_creator.go:188)
- [ ] Refactor `applyTicketHolderName` to avoid mutating pass fields in place (src/pkg/wallet/apple/embedded_creator.go:207)
- [ ] Refactor `updateFieldValue` to avoid mutating the provided fields slice (src/pkg/wallet/apple/embedded_creator.go:276)
- [ ] Remove mutation from batch entrypoint config validation flow (src/cmd/batch/main.go:35)
- [ ] Verify the `defaultTicketStatus` constant provides value or remove it (src/pkg/batch/wallet.go:26)
- [ ] Restore validator usage without stack overflow in batch wallet syncer (src/pkg/batch/wallet.go:59)
- [ ] Replace the inline `newTicketTailorTicketFetcher` wiring with injected dependency (src/pkg/batch/wallet.go:88)
- [ ] Rework the batch validator initialization to avoid the TODO (src/pkg/batch/wallet.go:94)
- [ ] Mark produced passes based on generated artifacts rather than source tickets (src/pkg/batch/wallet.go:140)
- [ ] Remove the legacy `newTicketTailorTicketFetcher` helper (src/pkg/batch/wallet.go:231)
- [ ] Improve the global validator initialization in the default Apple pass creator (src/pkg/wallet/apple/default_creator.go:49)
