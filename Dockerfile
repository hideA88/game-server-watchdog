# Build stage
FROM golang:1.24-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates

# ワークディレクトリを設定
WORKDIR /app

# 依存関係をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o game-server-watchdog ./cmd/watchdog/main.go

# Runtime stage
FROM alpine:latest

# 必要なパッケージをインストール
# Docker APIを使用するため、docker-cliとdocker-composeは不要
RUN apk --no-cache add ca-certificates tzdata \
  # ホストシステム監視用のツール
  procps sysstat

# 非rootユーザーを作成
RUN addgroup -g 1000 -S watchdog && \
    adduser -u 1000 -S watchdog -G watchdog

# 作業ディレクトリを作成
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/game-server-watchdog .

# 実行権限を付与
RUN chmod +x game-server-watchdog

# 所有権を変更
RUN chown -R watchdog:watchdog /app

# 非rootユーザーに切り替え
USER watchdog

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep game-server-watchdog || exit 1

# エントリーポイント
ENTRYPOINT ["./game-server-watchdog"]
