ARG GO_VERSION=1.24

FROM golang:${GO_VERSION} AS builder

WORKDIR /src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /src/bin/out ./cmd/batch

FROM alpine:3.20

RUN apk add --no-cache curl ca-certificates tzdata


# Latest releases available at https://github.com/aptible/supercronic/releases
ENV SUPERCRONIC_URL=https://github.com/aptible/supercronic/releases/download/v0.2.38/supercronic-linux-amd64 \
    SUPERCRONIC_SHA1SUM=bc072eba2ae083849d5f86c6bd1f345f6ed902d0 \
    SUPERCRONIC=supercronic-linux-amd64

RUN curl -fsSLO "$SUPERCRONIC_URL" \
    && echo "${SUPERCRONIC_SHA1SUM}  ${SUPERCRONIC}" | sha1sum -c - \
    && chmod +x "$SUPERCRONIC" \
    && mv "$SUPERCRONIC" "/usr/local/bin/${SUPERCRONIC}" \
    && ln -s "/usr/local/bin/${SUPERCRONIC}" /usr/local/bin/supercronic


WORKDIR /app

COPY --from=builder /src/bin/out /app/out

RUN mkdir -p /app/cron.d \
    && printf '*/5 * * * * /app/out\n' > /app/cron.d/wallet.cron
