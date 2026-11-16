# Run `make` or `make help` to list targets.

.DEFAULT_GOAL := help

.PHONY: help up down reset restart logs ps open password

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-9s\033[0m %s\n", $$1, $$2}'

# First run: create .env from .env.example with a random Grafana password.
.env:
	@sed "s/^GRAFANA_ADMIN_PASSWORD=.*/GRAFANA_ADMIN_PASSWORD=$$(head -c 18 /dev/urandom | base64 | tr -d '/+=')/" .env.example > .env
	@echo "Created observability/.env with a random Grafana password: \`make -C observability password\` shows it."

up: .env ## Start collector, Tempo, Loki, Prometheus and Grafana (http://localhost:3000)
	@mkdir -p ../backend/logs
	docker compose up -d --wait --remove-orphans
	@echo "Grafana: http://localhost:3000 (user admin; password: make -C observability password)"

down: ## Stop the stack (collected data is kept)
	docker compose down

reset: ## Stop the stack and DELETE all collected traces, metrics and logs
	docker compose down -v

restart: ## Restart Grafana and the collector (after editing their config files)
	docker compose restart grafana otel-collector

logs: ## Follow the stack's own logs (to debug the stack itself)
	docker compose logs -f --tail=50

ps: ## Show the stack's containers and their state
	docker compose ps

open: ## Open Grafana in the browser
	xdg-open http://localhost:3000 >/dev/null 2>&1 || echo "open http://localhost:3000"

password: .env ## Print the Grafana admin password
	@grep '^GRAFANA_ADMIN_PASSWORD=' .env | cut -d= -f2-
