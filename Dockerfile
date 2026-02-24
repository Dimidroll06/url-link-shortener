FROM golang:1.26-alpine3.23 AS builder

RUN apk add --no-cache git ca-certificates tzdata
RUN adduser -D -g '' appuser

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test ./... -v
RUN go build -ldflags="-s -w -extldflags '-static'" -o bin/server ./cmd/server

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/server .

USER appuser

# Должен совпадать с SERVER_PORT в .env
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]