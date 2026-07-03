# Stage 1: Build the Go binary and goose CLI
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.22.1

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4 && \
    swag init --parseDependency --parseInternal --parseDepth 3 --overridesFile .swaggo -g cmd/api/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/shopping-platform-api ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/shopping-platform-api /app/shopping-platform-api
COPY --from=builder /app/docs/swagger.json /app/docs/swagger.json
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY internal/migrations /app/internal/migrations
COPY scripts/docker-entrypoint.sh /app/docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
