.DEFAULT_GOAL := help

BIN := bin


.PHONY: help
help: ## List available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

## ---------------------------------------------------------------- build & run

.PHONY: build-nextjs
build-nextjs: ## Create nextjs export
	npm install --prefix frontend; \
	NEXT_TELEMETRY_DISABLED=1 npm run build --prefix frontend

.PHONY: build
build: build-nextjs ## Build the app
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
