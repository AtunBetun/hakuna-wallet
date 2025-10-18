ARG GO_VERSION=1.24

FROM golang:${GO_VERSION} AS builder

WORKDIR /src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /src/bin/out ./cmd/batch

RUN mkdir -p /src/runtime/data /src/runtime/tickets

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=builder /src/bin/out /app/out
COPY --from=builder --chown=nonroot:nonroot /src/runtime/data /app/data
COPY --from=builder --chown=nonroot:nonroot /src/runtime/tickets /app/tickets
COPY --chown=nonroot:nonroot certs /app/certs

ENV DATA_DIR=/app/data
ENV TICKETS_DIR=/app/tickets

ENTRYPOINT ["/app/out"]
