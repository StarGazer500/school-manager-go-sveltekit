# Run `make` or `make help` to list targets.

PORT    ?= 8080
DB_PATH ?= school.db
BIN     := bin/api

.DEFAULT_GOAL := help

.PHONY: help run build test vet fmt tidy check clean reset-db

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

run: ## Start the API (PORT=8080, DB_PATH=school.db)
	PORT=$(PORT) DB_PATH=$(DB_PATH) go run ./cmd/api

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

reset-db: ## Delete the SQLite database (recreated empty on next run)
	rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal
