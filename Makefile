.PHONY: help
help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

# Goプロジェクト用のコマンド
.PHONY: build
build: ## Build the watchdog binary
	go build -o bin/watchdog ./cmd/watchdog

.PHONY: run
run: build ## Build and run the watchdog
	./bin/watchdog

.PHONY: test
test: ## Run tests
	go tool gotestsum --format testname -- -v -race -cover ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: test-short
test-short: ## Run short tests
	go test -short ./...

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/
	rm -f coverage.out

.PHONY: lint
lint: ## Run golangci-lint
	go tool golangci-lint run

.PHONY: format
format: ## Format code
	go tool goimports -w .
	go fmt ./...

.PHONY: tidy
tidy: ## Tidy go.mod file
	go mod tidy

.PHONY: deps
deps: ## Update all dependencies
	go get -u ./...


