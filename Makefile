.DEFAULT_GOAL := help

BIN := bin


.PHONY: help
help: ## List available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

## ---------------------------------------------------------------- build & run

.PHONY: generate
generate: ## generate code files (sqlc)
	go tool sqlc generate

.PHONY: build-nextjs
build-nextjs: ## Create nextjs export
	npm install --prefix frontend; \
	NEXT_TELEMETRY_DISABLED=1 npm run build --prefix frontend

.PHONY: build
build: generate build-nextjs ## Build the app
	go build -o $(BIN)/jao ./cmd/jao

.PHONY: run
run: ## Run the app on the host
	go run ./... serve

.PHONY: check-no-lint
check-no-lint: ## Run checks without lint: tidy, vet, test - for CI
	go mod tidy -diff
	go vet ./...
	go tool gotestsum --format dots-v2 --format-hide-empty-pkg -- ./...

.PHONY: check
check: check-no-lint ## Run checks: tidy, vet, test, lint
	golangci-lint run

.PHONY: test-integration
test-integration: ## Run integration tests (needs Docker)
	go tool gotestsum --format dots-v2 --format-hide-empty-pkg -- -tags integration -count=1 ./...

## ---------------------------------------------------------------- db

.PHONY: migrate
migrate: ## db: apply pending migrations
	go run ./cmd/jao/... migrate up

.PHONY: migrate-down
migrate-down: ## db: roll back most recent migration
	go run ./cmd/jao/... migrate down

.PHONY: migrate-version
migrate-version: ## db: get db migration version
	go run ./cmd/jao/... migrate version

.PHONY: migrate-new
migrate-new: ## db: create new migration file
	@test -n "$(name)" || { echo "usage: make migration name=add_something" >&2; exit 1; }
	goose -dir db/migrations/sql create $(name) sql
