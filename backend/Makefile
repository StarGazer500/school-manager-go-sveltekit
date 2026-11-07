# Run `make` or `make help` to list targets.

PORT         ?= 8080
DATABASE_URL ?= postgres://school:school@localhost:5433/school?sslmode=disable
BIN          := bin/api

# pgrepo lives in a private GitHub repo. GOPRIVATE makes Go fetch it with git
# (your SSH key) and skip the public checksum database, which can't see it.
export GOPRIVATE := github.com/StarGazer500

.DEFAULT_GOAL := help

.PHONY: help run build test vet fmt tidy check clean db-up db-down db-reset db-shell \
	migrate-up migrate-down migrate-status migrate-new

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

run: db-up ## Start Postgres (if needed) and the API
	PORT=$(PORT) DATABASE_URL='$(DATABASE_URL)' go run ./cmd/api

build: ## Compile to bin/api
	go build -o $(BIN) ./cmd/api

test: ## Run tests
	go test ./...

vet: ## Static checks
	go vet ./...

fmt: ## Format all Go files
	gofmt -w .

tidy: ## Sync go.mod/go.sum with imports
	go mod tidy

check: vet test ## vet + test

clean: ## Remove build output
	rm -rf bin

db-up: ## Start Postgres in Docker and wait until it's ready
	docker compose up -d --wait db

db-down: ## Stop Postgres (data is kept)
	docker compose down

db-reset: ## Stop Postgres and DELETE all its data
	docker compose down -v

db-shell: ## Open psql on the database
	docker compose exec db psql -U school -d school

# Migrations live in migrations/ as numbered .sql files. The API also runs
# migrate-up by itself at startup, so these are for checking and undoing.
migrate-up: db-up ## Apply all new migrations
	DATABASE_URL='$(DATABASE_URL)' go run ./cmd/migrate up

migrate-down: db-up ## Undo the most recent migration
	DATABASE_URL='$(DATABASE_URL)' go run ./cmd/migrate down

migrate-status: db-up ## Show which migrations have run
	DATABASE_URL='$(DATABASE_URL)' go run ./cmd/migrate status

migrate-new: ## Create a migration: make migrate-new name=add_phone_to_students
	@test -n "$(name)" || { echo 'usage: make migrate-new name=what_it_does'; exit 2; }
	go run ./cmd/migrate new $(name)
