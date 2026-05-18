include .env
export

MIGRATIONS_DIR = internal/infrastructure/database/migrations
DATABASE_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
LATEST_VERSION = $(shell ls -1 $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null | sed 's/.*\///' | sort -n | tail -1 | grep -oE '^[0-9]+')

.PHONY: swagger migrate run

swagger:
	swag init -g cmd/main.go -o docs

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(LATEST_VERSION)

run: swagger migrate-up
	go run ./cmd/main.go
