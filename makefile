MIGRATIONS_DIR := migrations

POSTGRES_CONTAINER := local-postgres
POSTGRES_USER := brief_user
POSTGRES_PASSWORD := brief_password
POSTGRES_DB := brief_db

TIMESTAMP := $(shell date +"%Y%m%d%H%M%S")
NAME ?= migration

UP_FILE := $(MIGRATIONS_DIR)/$(TIMESTAMP)_$(NAME).up.sql
DOWN_FILE := $(MIGRATIONS_DIR)/$(TIMESTAMP)_$(NAME).down.sql

.PHONY: create-migration migrate
create-migration:
	@mkdir -p $(MIGRATIONS_DIR)
	@touch $(UP_FILE) $(DOWN_FILE)
	@echo "Created migration files:"
	@echo "  $(UP_FILE)"
	@echo "  $(DOWN_FILE)"

MIGRATIONS_DIR := migrations
DATABASE_URL := "postgres://brief_user:brief_password@localhost:5432/brief_db?sslmode=disable"

.PHONY: migrate-up migrate-down

migrate-up:
	@echo "Running all pending migrations (up)..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) up
	@echo "Migrations up complete."

migrate-down:
	@echo "Rolling back last migration (down)..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) down 1
	@echo "Migration down complete."
