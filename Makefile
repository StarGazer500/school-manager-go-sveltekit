# Convenience targets for the whole project. Each folder also has its own Makefile.

.DEFAULT_GOAL := help

.PHONY: help dev api web check obs-up obs-down obs-reset obs-open

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

dev: ## Start backend and frontend together (Ctrl+C stops both)
	$(MAKE) -j2 api web

api: ## Start only the backend
	$(MAKE) -C backend run

web: ## Start only the frontend
	$(MAKE) -C frontend dev

check: ## Check both halves
	$(MAKE) -C backend check
	$(MAKE) -C frontend check

obs-up: ## Start the observability stack (Grafana on http://localhost:3000)
	$(MAKE) -C observability up

obs-down: ## Stop the observability stack (data is kept)
	$(MAKE) -C observability down

obs-reset: ## Stop the observability stack and delete its data
	$(MAKE) -C observability reset

obs-open: ## Open Grafana in the browser
	$(MAKE) -C observability open
