# Run `make` or `make help` to list targets.

SHELL   := /bin/bash
API_URL ?= http://localhost:8080/api/v1
NVM_DIR ?= $(HOME)/.nvm

# nvm is a shell function, not a program, so it has to be loaded in each
# recipe. When nvm is installed, this switches to the Node version in .nvmrc
# first; otherwise whatever `node` is on PATH is used.
NODE := $(if $(wildcard $(NVM_DIR)/nvm.sh),source $(NVM_DIR)/nvm.sh && nvm use --silent >/dev/null &&,)

.DEFAULT_GOAL := help

.PHONY: help install dev build preview check clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

install: ## Install npm dependencies
	$(NODE) npm install

dev: node_modules ## Start the dev server on http://localhost:5173
	$(NODE) API_URL=$(API_URL) npm run dev

build: node_modules ## Production build
	$(NODE) npm run build

preview: build ## Serve the production build locally
	$(NODE) API_URL=$(API_URL) npm run preview

check: node_modules ## Type-check (svelte-check)
	$(NODE) npm run check

clean: ## Remove build output
	rm -rf .svelte-kit build

# Installs dependencies automatically the first time.
node_modules: package.json
	$(NODE) npm install
	@touch node_modules
