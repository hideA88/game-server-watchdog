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
test: ## Run all tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/
	rm -f coverage.out

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: tidy
tidy: ## Tidy go.mod file
	go mod tidy

.PHONY: deps
deps: ## Update all dependencies
	go get -u ./...

.PHONY: install-linter
install-linter: ## Install golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

