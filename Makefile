COMPOSE ?= docker compose

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://neobit:neobit@localhost:5432/neobit?sslmode=disable
MIGRATIONS_DIR=internal/db/migrations

.PHONY: db-up db-down db-logs migrate-up migrate-down migrate-status migrate-create psql

db-up:
	$(COMPOSE) up -d db

db-down:
	$(COMPOSE) down -v

db-logs:
	$(COMPOSE) logs -f db

migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

migrate-create:
	goose -dir $(MIGRATIONS_DIR) create new_migration sql

psql:
	PGPASSWORD=neobit psql -h localhost -p 5432 -U neobit -d neobit
