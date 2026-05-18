FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bank_app ./cmd/main.go

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /bank_app .
COPY --from=builder /app/internal/infrastructure/database/migrations ./internal/infrastructure/database/migrations

EXPOSE 8080

CMD ["./bank_app"]
