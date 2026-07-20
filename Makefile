GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= postgres://tourik:tourik@localhost:5432/tourik?sslmode=disable
MIGRATIONS_DIR := migrations

.PHONY: build run
build:
	go build -o bin/tourik ./cmd/tourik

run:
	go run ./cmd/tourik

.PHONY: migrate-create migrate-up migrate-down migrate-status
migrate-create:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
		go tool goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
		go tool goose -dir $(MIGRATIONS_DIR) up

migrate-down:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
		go tool goose -dir $(MIGRATIONS_DIR) down

migrate-status:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
		go tool goose -dir $(MIGRATIONS_DIR) status