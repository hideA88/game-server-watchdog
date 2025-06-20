.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: format
format: ## コードをフォーマット
	@echo "Running goimports..."
	@goimports -w .
	@echo "Running go fmt..."
	@go fmt ./...

.PHONY: lint
lint: ## lintを実行
	@echo "Running golangci-lint..."
	@golangci-lint run

.PHONY: test
test: ## テストを実行
	@echo "Running tests..."
	@go tool gotestsum --format testname -- -v -race -cover ./...

.PHONY: test-coverage
test-coverage: ## カバレッジ付きでテストを実行
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: coverage-filtered
coverage-filtered: ## カバレッジ計測（.coverageignoreに記載されたファイルを除外）
	@echo "Running tests with filtered coverage..."
	@go test -coverprofile=coverage.out.tmp ./...
	@cat coverage.out.tmp | grep -v -f .coverageignore > coverage.out || cp coverage.out.tmp coverage.out
	@echo "Coverage (excluding ignored files):"
	@go tool cover -func=coverage.out | grep -E "^total:"

.PHONY: build
build: ## バイナリをビルド
	@echo "Building binary..."
	@go build -o bin/game-server-watchdog ./cmd/watchdog/main.go

.PHONY: install
install: ## バイナリをインストール
	@echo "Installing binary..."
	@go install ./cmd/watchdog

.PHONY: run
run: ## ボットを起動
	@echo "Starting bot..."
	@go run ./cmd/watchdog/main.go

.PHONY: clean
clean: ## ビルド成果物を削除
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.out.tmp coverage.html
	@go clean -testcache

.PHONY: deps
deps: ## 依存関係を更新
	@echo "Updating dependencies..."
	@go mod download
	@go mod tidy

.PHONY: release-dry-run
release-dry-run: ## リリースのドライラン（GoReleaser）
	@echo "Running release dry-run..."
	@goreleaser release --snapshot --clean

.PHONY: release-local
release-local: ## ローカルでリリースビルドを実行
	@echo "Building release locally..."
	@goreleaser build --snapshot --clean

.PHONY: tag
tag: ## Git タグを作成（例: make tag VERSION=v1.0.0）
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set. Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag $(VERSION) created. Push with: git push origin $(VERSION)"

.PHONY: docker-build
docker-build: ## Dockerイメージをビルド
	@echo "Building Docker image..."
	@docker build -t game-server-watchdog:latest .

.PHONY: docker-run
docker-run: ## Dockerコンテナを起動
	@echo "Running Docker container..."
	@docker run --rm -it \
		--env-file .env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD)/docker-compose.yml:/root/docker-compose.yml:ro \
		game-server-watchdog:latest

# 開発環境用のターゲット
.PHONY: dev-up
dev-up: ## 開発環境を起動（ゲームサーバー + Watchdog）
	@echo "Starting development environment..."
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml up -d --build

.PHONY: dev-down
dev-down: ## 開発環境を停止
	@echo "Stopping development environment..."
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml down

.PHONY: dev-logs
dev-logs: ## Watchdogのログを表示
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml logs -f watchdog

.PHONY: dev-rebuild
dev-rebuild: ## Watchdogを再ビルドして再起動
	@echo "Rebuilding and restarting watchdog..."
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml build watchdog
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml up -d watchdog

.PHONY: dev-restart
dev-restart: ## Watchdogを再起動（ビルドなし）
	@echo "Restarting watchdog..."
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml restart watchdog

.PHONY: dev-status
dev-status: ## 開発環境のステータスを表示
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml ps

.PHONY: dev-clean
dev-clean: ## 開発環境を完全にクリーンアップ
	@echo "Cleaning up development environment..."
	@docker compose -p dev -f infra/dev/docker-compose.dev.yml down -v