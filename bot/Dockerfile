FROM golang:1.24.6 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o hearsay ./cmd/hearsay/main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y sqlite3 libsqlite3-dev gcc tzdata && rm -rf /var/lib/apt/lists/*


COPY --from=builder /app/hearsay .
COPY config.yaml .