GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= postgres://tourik:tourik@localhost:5432/tourik?sslmode=disable
MIGRATIONS_DIR := migrations

.PHONY: build run test test-integration compose-up rebuild-app migrate-create migrate-up migrate-down migrate-status swag

build:
	go build -o bin/tourik ./cmd/tourik

run: build
	./bin/tourik

test:
	go test ./...

# Spins up real Postgres/Kafka containers via testcontainers-go, so it
# requires a running Docker daemon.
test-integration:
	go test -tags=integration ./test/integration/... -v

swag:
	swag init -g cmd/tourik/main.go  -d ./,./cmd/tourik,./internal

compose-up:
	docker compose up -d

rebuild-app:
	docker compose up -d --build app

mig-create:
	go tool goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" create $(name) sql

mig-up:
	go tool goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

mig-down:
	go tool goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

mig-status:
	go tool goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status