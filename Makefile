GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= postgres://tourik:tourik@localhost:5432/tourik?sslmode=disable
MIGRATIONS_DIR := migrations

.PHONY: build compose-up rebuild-app migrate-create migrate-up migrate-down migrate-status swag

build:
	go build -o bin/tourik ./cmd/tourik

swag:
	swag init -g cmd/tourik/main.go  -d ./,./cmd/tourik,./internal

compose-up:
	docker compose up -d

rebuild-app:
	docker compose up -d --build app

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