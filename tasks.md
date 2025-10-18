# Tasks

## Ticket lifecycle
- [ ] Stand up Postgres database for ticket issuance metadata
- [ ] Define schema to record generated tickets (ticket id, issued at, fulfillment status)
- [ ] Add persistence layer to write/read from Postgres during ticket generation

## Ticket assets
- [ ] Integrate the designer-provided ticket template into the generator

## Email delivery
- [ ] Build email workflow to send the generated ticket

## Deployment
- [ ] Prepare Fly.io configuration (Dockerfile, fly.toml, secrets)
- [ ] Automate github deploy pipeline to Fly.io
- [ ] Set up semantic versioning workflow for releases
- [x] Gate Fly deploy workflow on successful CI run
- [x] Materialize certificate secrets into runtime files at startup
