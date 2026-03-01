.PHONY: help migrateup migratedown migrateforce

help: ## Show available commands
	@echo "Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

migrateup: ## Run database migrations up (loads .env if present)
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$$DSN" ]; then echo "Error: set DSN in .env"; exit 1; fi; \
	migrate -path db/migrations -database "mysql://$$DSN" -verbose up

migratedown: ## Run database migrations down (loads .env if present)
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$$DSN" ]; then echo "Error: set DSN in .env"; exit 1; fi; \
	migrate -path db/migrations -database "mysql://$$DSN" -verbose down

migrateforce: VERSION ?= 1
migrateforce: ## Fix dirty DB: make migrateforce [VERSION=1] (clears dirty; use 1 if schema exists)
	@[ -f .env ] && set -a && . ./.env && set +a; \
	if [ -z "$$DSN" ]; then echo "Error: set DSN in .env"; exit 1; fi; \
	migrate -path db/migrations -database "mysql://$$DSN" force $(VERSION)
